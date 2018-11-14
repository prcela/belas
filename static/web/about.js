class AboutViewController extends ViewController {
	constructor(node) {
		super(node)
		node.className = "AboutViewController"
		this.title = "About"

		var divContactUs = document.createElement("div")
		divContactUs.className = "MenuItem"
		
		var a = document.createElement("a")
		a.href = "mailto:yamb.igre@gmail.com"
		a.innerText = "Contact us"
		divContactUs.appendChild(a)
		this.node.appendChild(divContactUs)
	}
	show() {
		super.show()
	}
}