class PlayerStat {
	constructor() {
		this.node = document.createElement("div")
		this.node.className = "Toolbar"
		document.listeners["onPlayerStat"].push(this)
	}
	show() {
		this.node.innerHTML = ""
		if (PlayerStat.player != undefined) {
			var divTable = document.createElement("table")
			var divR0 = document.createElement("tr")
			var divT0 = document.createElement("td")
			divT0.appendChild(document.createTextNode(PlayerStat.player.alias))
			divTable.appendChild(divR0);divR0.appendChild(divT0);
			var divR1 = document.createElement("tr")
			var divT1 = document.createElement("td")
			divT1.appendChild(document.createTextNode(PlayerStat.player.pretzels + " 🥨  " + PlayerStat.player.diamonds + " 💎  " + stars6(PlayerStat.avgScore(6)) + " ⭐️"))
			divTable.appendChild(divR0);divR0.appendChild(divT0);
			divTable.appendChild(divR1);divR1.appendChild(divT1);
			this.node.appendChild(divTable)	
		}
	}
	onPlayerStat(e) {
		this.show()
	}

	static avgScore(diceNum) {
		var items = this.stat_items.filter(function(item) {
			return item.dice_num == diceNum
		})
		if (items.length == 0) {
			return 0
		}
		var sum = 0
		for( var i = 0; i < items.length; i++ ) {
    		sum += item.score
		}
		var avg = sum/items.length
		return avg
	}
}