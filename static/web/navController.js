class NavigationController {
	constructor(node,rootVC) {
		this.node = node
		this.node.className = "NavigationController"
		rootVC.navigationController = this
		this.items = [rootVC]
		this.node.appendChild(rootVC.node)
		rootVC.show()
	}
	push(viewController) {
	}
}