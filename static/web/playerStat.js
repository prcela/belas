class PlayerStat {
	constructor() {
		this.node = document.createElement("div")
		this.node.className = "Toolbar"
		document.listeners["onPlayerStat"].push(this)
	}
	show() {
		this.node.innerHTML = ""
		if (PlayerStat.player != undefined) {
			var divAlias = document.createElement("p")
			divAlias.appendChild(document.createTextNode(PlayerStat.player.alias))
			this.node.appendChild(divAlias)	
		}
	}
	onPlayerStat(e) {
		this.show()
	}
}