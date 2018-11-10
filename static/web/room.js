var RoomSection = {oneVsAll:0, create:1, freeTables:2, freeTournaments:3, freePlayers:4, tables:5, tournaments:6}
Object.freeze(RoomSection)

class Room {
  constructor(node) {
  	this.node = node
  	this.sections = []
  	document.roomInfoListeners.push(this)
  	wsAPI.roomInfo()
  }
  show() {
  	this.node.innerHTML = ""
  	this.appendSection("Free players")
  	var freePlayers = Object.values(Room.players)
  	for (var i = freePlayers.length - 1; i >= 0; i--) {
  		this.appendPlayer(freePlayers[i])
  	}
  }
  appendSection(title) {
  	var newDiv = document.createElement("div")
  	newDiv.className = "TableSection"
	var newContent = document.createTextNode(title)
	newDiv.appendChild(newContent)
    this.node.appendChild(newDiv)
  }
  appendPlayer(p) {
  	var newDiv = document.createElement("div")
  	newDiv.className = "Player"
  	newDiv.appendChild(document.createTextNode(p.alias))
  	this.node.appendChild(newDiv)
  }
  onRoomInfo(e) {
  	console.log("onRoomInfo event received")
  	this.show()
  }
}

Room.players = []