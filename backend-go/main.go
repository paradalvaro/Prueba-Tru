package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"

	"os"

	"github.com/dgraph-io/dgo/v200"
	"github.com/dgraph-io/dgo/v200/protos/api"
)

//var routes = flag.Bool("routes", false, "Generate router documentation")

func main() {
	flag.Parse()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*", "localhost"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("root1."))
	})

	// RESTy routes for "programs" resource
	r.Route("/program", func(r chi.Router) {
		r.With(paginate).Get("/", ListPrograms)
		r.Post("/", NewProgram) // POST

		r.Route("/{programID}", func(r chi.Router) {
			r.Get("/", GetProgram)
		})
	})
	http.ListenAndServe(":3333", r)
}

// paginate is a stub, but very possible to implement middleware logic
// to handle the request params for handling a paginated request.
func paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// just a stub.. some ideas are to look at URL query params for something like
		// the page number, or the limit, and send a query cursor down the chain
		next.ServeHTTP(w, r)
	})
}

type CancelFunc func()

func getDgraphClient() (*dgo.Dgraph, CancelFunc) {
	conn, err := dgo.DialSlashEndpoint("https://blue-surf-590284.us-east-1.aws.cloud.dgraph.io/graphql", "MmU3OGI2OGFhNDY2NTQxMmE5MWE3MDkxZTYwNDIwZDU=")
	if err != nil {
		log.Fatal("While trying to dial gRPC")
	}
	dc := api.NewDgraphClient(conn)
	dg := dgo.NewDgraphClient(dc)
	return dg, func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error while closing connection:%v", err)
		}
	}
}

type Node struct {
	Id           string      `json:"id,omitempty"`
	Name         string      `json:"name,omitempty"`
	Data         interface{} `json:"data,omitempty"`
	Inputs       interface{} `json:"inputs,omitempty"`
	Outputs      interface{} `json:"outputs,omitempty"`
	Code         string      `json:"code,omitempty"`
	InternalCode string      `json:"internalcode,omitempty"`
}

type NodeList struct {
	Nodes []Node
}

type Program struct {
	Uid        string `json:"uid,omitempty"`
	Name       string `json:"Program.name,omitempty"`
	Code       string `json:"Program.code,omitempty"`
	StringCode string `json:"stringcode,omitempty"`
	Result     string `json:"result,omitempty"`
}

func ListPrograms(w http.ResponseWriter, r *http.Request) {

	dg, cancel := getDgraphClient()

	defer cancel()

	ctx := context.Background()
	q := `{
			all (func: has(Program.name) ){
				uid
				Program.name				
			}
		}`

	txn := dg.NewTxn()
	defer txn.Discard(ctx)

	resp, _ := txn.Query(ctx, q)
	//fmt.Println(string(resp.Json))

	type ListPro struct {
		Query []*Program `json:"all,omitempty"`
	}

	var aux ListPro

	json.Unmarshal(resp.Json, &aux)
	//fmt.Println(aux)

	if err := render.RenderList(w, r, NewProgramListResponse(aux.Query)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func NewProgram(w http.ResponseWriter, r *http.Request) {

	request := map[string]string{}
	json.NewDecoder(r.Body).Decode(&request)

	article := Program{
		Uid:  request["uid"],
		Name: request["name"],
		Code: request["code"],
	}

	dg, cancel := getDgraphClient()
	defer cancel()

	ctx := context.Background()
	mu := &api.Mutation{
		CommitNow: true,
	}

	pb, err := json.Marshal(article)
	if err != nil {
		log.Fatal(err)
	}
	mu.SetJson = pb
	_, err = dg.NewTxn().Mutate(ctx, mu)
	if err != nil {
		log.Fatal(err)
	}

	if request["exec"] == "true" {
		fmt.Println("Execute Code")
		var internalStringCode string
		article.StringCode, internalStringCode = executeCodeProgram(request["code"])
		code := "if(True):\n\t" + strings.ReplaceAll(internalStringCode, "\n", "\n\t") + "\n"

		err := os.WriteFile("CodetoExecute.py", []byte(code), 0666)
		if err != nil {
			println(err.Error())
			return
		}

		cmd := exec.Command("python", "Execute.py")
		out, err := cmd.Output()
		if err != nil {
			println(err.Error())
			return
		}
		article.Result = string(out)
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, NewProgramResponse(&article))
}

func executeCodeProgram(code string) (string, string) {

	var nodes NodeList

	var result map[string]interface{}
	json.Unmarshal([]byte(code), &result)
	aux := result["drawflow"].(map[string]interface{})["Home"].(map[string]interface{})["data"].(map[string]interface{})

	for key, value := range aux {
		node := Node{
			Id: key,
		}
		for key1, value1 := range value.(map[string]interface{}) {
			switch key1 {
			case "outputs":
				node.Outputs = value1.(map[string]interface{})["output_1"]
			case "name":
				node.Name = value1.(string)
			case "inputs":
				node.Inputs = value1.(map[string]interface{})
			case "data":
				node.Data = value1.(map[string]interface{})
			}
		}
		nodes.Nodes = append(nodes.Nodes, node)
	}

	for key, value := range nodes.Nodes {
		/*sort.Strings(nodesIdChecked)
		index := sort.SearchStrings(nodesIdChecked, value.Id)
		if 0 < len(nodesIdChecked) && len(nodesIdChecked) < index && nodesIdChecked[index] == value.Id {
			fmt.Println("Repeti", value.Id, value.Name)
			//continue
		}
		fmt.Println(nodesIdChecked)*/

		if value.Name == "Add" {
			nodes.Nodes[key] = generateCodeAdd(value, nodes)
		} else if value.Name == "Assing" {
			nodes.Nodes[key] = generateCodeAssing(value, nodes)
		} else if value.Name == "For" {
			nodes.Nodes[key] = generateCodeFor(value, nodes)
		} else if value.Name == "Code" {
			nodes.Nodes[key] = generateCodeBlock(value, nodes)
		} else if value.Name == "If" {
			nodes.Nodes[key] = generateCodeIf(value, nodes)
		} else if value.Name == "Number" {
			value.Code = value.Data.(map[string]interface{})["namevalue"].(string)
			value.InternalCode = value.Code
			nodes.Nodes[key] = value
		}
	}
	fmt.Println(nodesIdChecked)
	/*
		//Add
		for key, value := range nodes.Nodes {
			if value.Name == "Add" {
				nodes.Nodes[key] = generateCodeAdd(value, nodes)
			}
		}
		//Assing
		for key, value := range nodes.Nodes {
			if value.Name == "Assing" {
				nodes.Nodes[key] = generateCodeAssing(value, nodes)
			}
		}
		//If
		for key, value := range nodes.Nodes {
			if value.Name == "If" {
				nodes.Nodes[key] = generateCodeIf(value, nodes)
			}
		}
		//For
		for key, value := range nodes.Nodes {
			if value.Name == "For" {
				nodes.Nodes[key] = generateCodeFor(value, nodes)
			}
		}
		//Code
		for key, value := range nodes.Nodes {
			if value.Name == "Code" {
				nodes.Nodes[key] = generateCodeBlock(value, nodes)
			}
		}
	*/
	stringCode, stringInternalCode := "", ""
	//no Outputs
	for _, value := range nodes.Nodes {
		auxOutMap := getOutputMap(value.Outputs)
		if len(auxOutMap) == 0 {
			fmt.Println(value.InternalCode)
			stringCode += value.Code + "\n"
			stringInternalCode += value.InternalCode + "\n"
		}
	}

	return stringCode, stringInternalCode

}

var nodesIdChecked []string

//GetCodes

func generateCodeIf(node Node, nodes NodeList) Node {

	/*sort.Strings(nodesIdChecked)
	index := sort.SearchStrings(nodesIdChecked, node.Id)
	if 0 < len(nodesIdChecked) && len(nodesIdChecked) < index && nodesIdChecked[index] == node.Id {
		fmt.Println("Repeti", node.Id, node.Name)
		//return node
	}*/
	nodesIdChecked = append(nodesIdChecked, node.Id)
	auxInMap := getInputMap(node.Inputs)
	node.Code = "if(" + node.Data.(map[string]interface{})["condition"].(string) + "):\n\t"
	node.InternalCode = node.Code

	if len(auxInMap) > 0 {
		for key, value := range nodes.Nodes {
			if value.Id == auxInMap["input_1"] {
				if value.Name == "Number" || value.Name == "Var" {
					nodesIdChecked = append(nodesIdChecked, value.Id)
					node.Code = node.Code + value.Data.(map[string]interface{})["namevalue"].(string)
					node.InternalCode = node.Code
				} else if value.Name == "Add" {
					nodes.Nodes[key] = generateCodeAdd(value, nodes)
					node.Code = node.Code + nodes.Nodes[key].Code
					node.InternalCode = node.Code
				} else if value.Name == "Assing" {
					nodes.Nodes[key] = generateCodeAssing(value, nodes)
					nodes.Nodes[key].InternalCode = strings.ReplaceAll(nodes.Nodes[key].InternalCode, "\n", "\n\t")
					node.InternalCode = node.Code + nodes.Nodes[key].InternalCode
					node.Code = node.Code + nodes.Nodes[key].Code
				} else if value.Name == "If" {
					nodes.Nodes[key] = generateCodeIf(value, nodes)
					nodes.Nodes[key].Code = strings.ReplaceAll(nodes.Nodes[key].Code, "\n", "\n\t")
					nodes.Nodes[key].InternalCode = strings.ReplaceAll(nodes.Nodes[key].InternalCode, "\n", "\n\t")
					node.InternalCode = node.Code + nodes.Nodes[key].InternalCode
					node.Code = node.Code + nodes.Nodes[key].Code
				} else if value.Name == "Code" {
					nodes.Nodes[key] = generateCodeBlock(value, nodes)
					nodes.Nodes[key].Code = strings.ReplaceAll(nodes.Nodes[key].Code, "\n", "\n\t")
					nodes.Nodes[key].InternalCode = strings.ReplaceAll(nodes.Nodes[key].InternalCode, "\n", "\n\t")
					node.InternalCode = node.Code + nodes.Nodes[key].InternalCode
					node.Code = node.Code + nodes.Nodes[key].Code
				} else if value.Name == "For" {
					nodes.Nodes[key] = generateCodeFor(value, nodes)
					nodes.Nodes[key].Code = strings.ReplaceAll(nodes.Nodes[key].Code, "\n", "\n\t")
					nodes.Nodes[key].InternalCode = strings.ReplaceAll(nodes.Nodes[key].InternalCode, "\n", "\n\t")
					node.InternalCode = node.Code + nodes.Nodes[key].InternalCode
					node.Code = node.Code + nodes.Nodes[key].Code
				}
			}
		}
		for key, value := range nodes.Nodes {
			if value.Id == auxInMap["input_2"] {
				node.Code += "\nelse:\n\t"
				node.InternalCode += "\nelse:\n\t"
				if value.Name == "Number" || value.Name == "Var" {
					nodesIdChecked = append(nodesIdChecked, value.Id)
					node.Code = node.Code + value.Data.(map[string]interface{})["namevalue"].(string)
					node.InternalCode += value.Data.(map[string]interface{})["namevalue"].(string)
				} else if value.Name == "Add" {
					nodes.Nodes[key] = generateCodeAdd(value, nodes)
					node.InternalCode = node.InternalCode + nodes.Nodes[key].InternalCode
					node.Code = node.Code + nodes.Nodes[key].Code
				} else if value.Name == "Assing" {
					nodes.Nodes[key] = generateCodeAssing(value, nodes)
					nodes.Nodes[key].InternalCode = strings.ReplaceAll(nodes.Nodes[key].InternalCode, "\n", "\n\t")
					node.InternalCode = node.InternalCode + nodes.Nodes[key].InternalCode
					node.Code = node.Code + nodes.Nodes[key].Code
				} else if value.Name == "If" {
					nodes.Nodes[key] = generateCodeIf(value, nodes)
					nodes.Nodes[key].Code = strings.ReplaceAll(nodes.Nodes[key].Code, "\n", "\n\t")
					nodes.Nodes[key].InternalCode = strings.ReplaceAll(nodes.Nodes[key].InternalCode, "\n", "\n\t")
					node.InternalCode = node.InternalCode + nodes.Nodes[key].InternalCode
					node.Code = node.Code + nodes.Nodes[key].Code
				} else if value.Name == "Code" {
					nodes.Nodes[key] = generateCodeBlock(value, nodes)
					nodes.Nodes[key].Code = strings.ReplaceAll(nodes.Nodes[key].Code, "\n", "\n\t")
					nodes.Nodes[key].InternalCode = strings.ReplaceAll(nodes.Nodes[key].InternalCode, "\n", "\n\t")
					node.InternalCode = node.InternalCode + nodes.Nodes[key].InternalCode
					node.Code = node.Code + nodes.Nodes[key].Code
				} else if value.Name == "For" {
					nodes.Nodes[key] = generateCodeFor(value, nodes)
					nodes.Nodes[key].Code = strings.ReplaceAll(nodes.Nodes[key].Code, "\n", "\n\t")
					nodes.Nodes[key].InternalCode = strings.ReplaceAll(nodes.Nodes[key].InternalCode, "\n", "\n\t")
					node.InternalCode = node.InternalCode + nodes.Nodes[key].InternalCode
					node.Code = node.Code + nodes.Nodes[key].Code
				}
			}
		}
	}
	return node

}

func generateCodeBlock(node Node, nodes NodeList) Node {

	/*sort.Strings(nodesIdChecked)
	index := sort.SearchStrings(nodesIdChecked, node.Id)
	if 0 < len(nodesIdChecked) && len(nodesIdChecked) < index && nodesIdChecked[index] == node.Id {
		fmt.Println("Repeti", node.Id, node.Name)
		//return node
	}*/

	nodesIdChecked = append(nodesIdChecked, node.Id)
	auxInMap := getInputMap(node.Inputs)
	node.Code = ""
	node.InternalCode = node.Code

	if len(auxInMap) > 0 {

		var keys []int
		for k := range auxInMap {
			i, _ := strconv.Atoi(strings.Split(k, "_")[1])
			keys = append(keys, i)
		}
		sort.Ints(keys)
		for _, k := range keys {
			value1 := auxInMap["input_"+strconv.Itoa(k)]
			for key, value := range nodes.Nodes {
				if value.Id == value1 {
					if value.Name == "Number" || value.Name == "Var" {
						nodesIdChecked = append(nodesIdChecked, value.Id)
						node.InternalCode = node.InternalCode + "\n" + value.Data.(map[string]interface{})["namevalue"].(string)
						node.Code = node.Code + "\n" + value.Data.(map[string]interface{})["namevalue"].(string)
					} else if value.Name == "Code" {
						nodes.Nodes[key] = generateCodeBlock(value, nodes)
						node.InternalCode = node.InternalCode + "\n" + nodes.Nodes[key].InternalCode
						node.Code = node.Code + "\n" + nodes.Nodes[key].Code
					} else if value.Name == "Add" {
						nodes.Nodes[key] = generateCodeAdd(value, nodes)
						node.InternalCode = node.InternalCode + "\n" + nodes.Nodes[key].InternalCode
						node.Code = node.Code + "\n" + nodes.Nodes[key].Code
					} else if value.Name == "Assing" {
						nodes.Nodes[key] = generateCodeAssing(value, nodes)
						node.InternalCode = node.InternalCode + "\n" + nodes.Nodes[key].InternalCode
						node.Code = node.Code + "\n" + nodes.Nodes[key].Code
					} else if value.Name == "For" {
						nodes.Nodes[key] = generateCodeFor(value, nodes)
						node.InternalCode = node.InternalCode + "\n" + nodes.Nodes[key].InternalCode
						node.Code = node.Code + "\n" + nodes.Nodes[key].Code
					} else if value.Name == "If" {
						nodes.Nodes[key] = generateCodeIf(value, nodes)
						node.InternalCode = node.InternalCode + "\n" + nodes.Nodes[key].InternalCode
						node.Code = node.Code + "\n" + nodes.Nodes[key].Code
					}
				}
			}
		}
	}
	return node

}

func generateCodeFor(node Node, nodes NodeList) Node {

	/*sort.Strings(nodesIdChecked)
	index := sort.SearchStrings(nodesIdChecked, node.Id)
	if 0 < len(nodesIdChecked) && len(nodesIdChecked) < index && nodesIdChecked[index] == node.Id {
		fmt.Println("Repeti", node.Id, node.Name)
		//return node
	}*/

	nodesIdChecked = append(nodesIdChecked, node.Id)
	auxInMap := getInputMap(node.Inputs)
	node.Code = "for x in range(" + node.Data.(map[string]interface{})["from"].(string) + "," + node.Data.(map[string]interface{})["until"].(string) + "):\n\t"
	node.InternalCode = node.Code
	if len(auxInMap) > 0 {
		for key, value := range nodes.Nodes {
			if value.Id == auxInMap["input_1"] {
				if value.Name == "Number" || value.Name == "Var" {
					nodesIdChecked = append(nodesIdChecked, value.Id)
					node.Code = node.Code + value.Data.(map[string]interface{})["namevalue"].(string)
					node.InternalCode = node.Code
				} else if value.Name == "Add" {
					//node.Code = node.Code + value.Code
					nodes.Nodes[key] = generateCodeAdd(value, nodes)
					node.Code = node.Code + nodes.Nodes[key].Code
					node.InternalCode = node.Code
				} else if value.Name == "Assing" {
					nodes.Nodes[key] = generateCodeAssing(value, nodes)
					nodes.Nodes[key].InternalCode = strings.ReplaceAll(nodes.Nodes[key].InternalCode, "\n", "\n\t")
					node.InternalCode = node.Code + nodes.Nodes[key].InternalCode
					node.Code = node.Code + nodes.Nodes[key].Code
				} else if value.Name == "If" {
					//node.Code = node.Code + strings.ReplaceAll(value.Code, "\n", "\n\t")
					nodes.Nodes[key] = generateCodeIf(value, nodes)
					nodes.Nodes[key].Code = strings.ReplaceAll(nodes.Nodes[key].Code, "\n", "\n\t")
					nodes.Nodes[key].InternalCode = strings.ReplaceAll(nodes.Nodes[key].InternalCode, "\n", "\n\t")
					node.InternalCode = node.Code + nodes.Nodes[key].InternalCode
					node.Code = node.Code + nodes.Nodes[key].Code
				} else if value.Name == "For" {
					nodes.Nodes[key] = generateCodeFor(value, nodes)
					nodes.Nodes[key].Code = strings.ReplaceAll(nodes.Nodes[key].Code, "\n", "\n\t")
					nodes.Nodes[key].InternalCode = strings.ReplaceAll(nodes.Nodes[key].InternalCode, "\n", "\n\t")
					node.InternalCode = node.Code + nodes.Nodes[key].InternalCode
					node.Code = node.Code + nodes.Nodes[key].Code
				} else if value.Name == "Code" {
					nodes.Nodes[key] = generateCodeBlock(value, nodes)
					nodes.Nodes[key].Code = strings.ReplaceAll(nodes.Nodes[key].Code, "\n", "\n\t")
					nodes.Nodes[key].InternalCode = strings.ReplaceAll(nodes.Nodes[key].InternalCode, "\n", "\n\t")
					node.InternalCode = node.Code + nodes.Nodes[key].InternalCode
					node.Code = node.Code + nodes.Nodes[key].Code
				}
			}
		}
	}

	return node

}

func generateCodeAssing(node Node, nodes NodeList) Node {

	/*sort.Strings(nodesIdChecked)
	index := sort.SearchStrings(nodesIdChecked, node.Id)
	if 0 < len(nodesIdChecked) && len(nodesIdChecked) < index && nodesIdChecked[index] == node.Id {
		fmt.Println("Repeti", node.Id, node.Name)
		//return node
	}*/

	nodesIdChecked = append(nodesIdChecked, node.Id)
	auxInMap := getInputMap(node.Inputs)
	node.Code = node.Data.(map[string]interface{})["varname"].(string) + " = "
	node.InternalCode = node.Code
	if len(auxInMap) > 0 {
		for key, value := range nodes.Nodes {
			if value.Id == auxInMap["input_1"] {
				if value.Name == "Number" || value.Name == "Var" {
					nodesIdChecked = append(nodesIdChecked, value.Id)
					node.Code = node.Code + value.Data.(map[string]interface{})["namevalue"].(string)
					node.InternalCode = node.Code
				} else if value.Name == "Add" {
					nodes.Nodes[key] = generateCodeAdd(value, nodes)
					node.Code = node.Code + nodes.Nodes[key].Code
					node.InternalCode = node.Code
				} /* else if value.Name == "Assing" {
					nodes.Nodes[key] = generateCodeAssing(value, nodes)
					//node.Code = node.Code + nodes.Nodes[key].Code
					node.Code = node.Code + nodes.Nodes[key].Code
				}*/
			}
		}
		node.InternalCode = node.InternalCode + "\n" + node.Data.(map[string]interface{})["varname"].(string) + "\n"
	}

	return node

}

func generateCodeAdd(node Node, nodes NodeList) Node {

	sort.Strings(nodesIdChecked)
	index := sort.SearchStrings(nodesIdChecked, node.Id)
	if 0 < len(nodesIdChecked) && len(nodesIdChecked) < index && nodesIdChecked[index] == node.Id {
		fmt.Println("Repeti", node.Id, node.Name)
		//return node
	}

	nodesIdChecked = append(nodesIdChecked, node.Id)
	auxInMap := getInputMap(node.Inputs)
	node.Code = node.Data.(map[string]interface{})["opvalue"].(string)
	node.InternalCode = node.Code
	if len(auxInMap) > 0 {
		for key, value := range nodes.Nodes {
			if value.Id == auxInMap["input_1"] {
				if value.Name == "Number" || value.Name == "Var" {
					nodesIdChecked = append(nodesIdChecked, value.Id)
					node.Code = value.Data.(map[string]interface{})["namevalue"].(string) + node.Code
					/* else if value.Name == "Assing" {
						//node.Code = value.Data.(map[string]interface{})["varname"].(string) + node.Code
						nodes.Nodes[key] = generateCodeAssing(value, nodes)
						node.Code = nodes.Nodes[key].Code + node.Code
					}*/
				} else if value.Name == "Add" {
					nodes.Nodes[key] = generateCodeAdd(value, nodes)
					node.Code = nodes.Nodes[key].Code + node.Code
				}
			} else if value.Id == auxInMap["input_2"] {
				//nodesIdChecked = append(nodesIdChecked, value.Id)
				if value.Name == "Number" || value.Name == "Var" {
					nodesIdChecked = append(nodesIdChecked, value.Id)
					node.Code = node.Code + value.Data.(map[string]interface{})["namevalue"].(string)
					/*} else if value.Name == "Assing" {
					//node.Code = node.Code + value.Data.(map[string]interface{})["varname"].(string)
					nodes.Nodes[key] = generateCodeAssing(value, nodes)
					node.Code = node.Code + nodes.Nodes[key].Code*/
				} else if value.Name == "Add" {
					nodes.Nodes[key] = generateCodeAdd(value, nodes)
					node.Code = node.Code + nodes.Nodes[key].Code
				}
			}
		}
		node.InternalCode = node.Code
	}

	return node

}

func getInputMap(inter interface{}) map[string]string {

	auxMap := make(map[string]string)
	aux1 := inter.(map[string]interface{})
	for key, value := range aux1 {
		for _, value1 := range value.(map[string]interface{}) {
			for _, value2 := range value1.([]interface{}) {
				for key1, value3 := range value2.(map[string]interface{}) {
					if key1 == "node" {
						auxMap[key] = value3.(string)
					}
				}
			}
		}
	}
	return auxMap
}

func getOutputMap(inter interface{}) map[string]string {
	auxMap := make(map[string]string)
	aux1 := inter.(map[string]interface{})["connections"]
	if aux1 != nil {
		for _, value := range aux1.([]interface{}) {
			for key1, value1 := range value.(map[string]interface{}) {
				auxMap[key1] = value1.(string)
			}
		}
	}
	return auxMap
}

type ListPro struct {
	Query []*Program `json:"all,omitempty"`
}

func GetProgram(w http.ResponseWriter, r *http.Request) {

	dg, cancel := getDgraphClient()

	defer cancel()

	p := Program{
		Uid: chi.URLParam(r, "programID"),
	}

	q := `query all($a: string) {
		all(func: uid($a)) {
		  Program.code
		}
	  }`

	ctx := context.Background()
	txn := dg.NewTxn()
	defer txn.Discard(ctx)

	resp, _ := txn.QueryWithVars(ctx, q, map[string]string{"$a": p.Uid})
	//fmt.Println(string(resp.Json))

	var aux ListPro

	json.Unmarshal(resp.Json, &aux)
	//fmt.Println(aux)

	p.Code = aux.Query[0].Code

	if err := render.Render(w, r, NewProgramResponse(&p)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

type ProgramRequest struct {
	*Program
	ProtectedID string `json:"id"` // override 'id' json to have more control
}

func (a *ProgramRequest) Bind(r *http.Request) error {
	// a.Article is nil if no Article fields are sent in the request. Return an
	// error to avoid a nil pointer dereference.
	/*if a.Program == nil {
		return errors.New("missing required Program fields.")
	}

	// a.User is nil if no Userpayload fields are sent in the request. In this app
	// this won't cause a panic, but checks in this Bind method may be required if
	// a.User or futher nested fields like a.User.Name are accessed elsewhere.

	// just a post-process after a decode..
	a.ProtectedID = ""                               // unset the protected ID
	a.Program.Name = strings.ToLower(a.Program.Name)*/ // as an example, we down-case
	return nil
}

type ProgramResponse struct {
	*Program
	// We add an additional field to the response here.. such as this
	// elapsed computed property
	Elapsed int64 `json:"elapsed"`
}

func NewProgramResponse(article *Program) *ProgramResponse {
	resp := &ProgramResponse{Program: article}

	return resp
}

func (rd *ProgramResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	rd.Elapsed = 10
	return nil
}

func NewProgramListResponse(articles []*Program) []render.Renderer {
	list := []render.Renderer{}
	for _, article := range articles {
		list = append(list, NewProgramResponse(article))
	}
	return list
}

// NOTE: as a thought, the request and response payloads for an Article could be the
// same payload type, perhaps will do an example with it as well.
// type ArticlePayload struct {
//   *Article
// }

//--
// Error response payloads & renderers
//--

// ErrResponse renderer type for handling all sorts of errors.
//
// In the best case scenario, the excellent github.com/pkg/errors package
// helps reveal information on the error, setting it on Err, and in the Render()
// method, using it to set the application-specific error code in AppCode.
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}
