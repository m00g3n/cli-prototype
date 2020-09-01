package workspace

const handlerPython = `def foo(event, context):
    return "hello world"`

var workspacePython = workspace{
	newTemplatedFile(handlerPython, "handler.py"),
}
