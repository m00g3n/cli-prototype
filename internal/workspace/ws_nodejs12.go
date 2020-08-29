package workspace

const handlerJs = `module.exports = {
    main: function (event, context) {
        return 'Hello Serverless'
    }
}`

const packageJSON = `{
  "name": "{{ .Name }}",
  "version": "0.0.1",
  "dependencies": {}
}`

var nodeJS12 = workspace{
	newFileTemplate(handlerJs, "handler.js"),
	newFileTemplate(packageJSON, "package.json"),
}
