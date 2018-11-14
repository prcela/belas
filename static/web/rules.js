class RulesViewController extends ViewController {
	constructor(node) {
    super(node)
    this.title = "Rules"
    node.className = "RulesViewController"
    var div = document.createElement("div")
    div.style.height = "fit-content"
    div.style.marginBottom = "120px"
    div.innerHTML='<object type="text/html" data="/static/rules/index_en.html" style="height: -webkit-fill-available;"></object>'
    node.appendChild(div)
  }
}