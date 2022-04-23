from code import InteractiveInterpreter
  
f = open('CodetoExecute.py','r')
code = f.read()
f.close()
# Using InteractiveInterpreter.runsource() method
InteractiveInterpreter().runsource(code)