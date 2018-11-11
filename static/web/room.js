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
  	this.appendSection("New game")
  	this.appendNewGame()
  	this.appendSection("Free tables")
  	var freeTables = Object.values(Room.tables).sort(function(t0,t1) {return sortStrings(t0.id,t1.id)})
  	for (var i = 0; i<freeTables.length; i++) {
  		this.appendTable(freeTables[i])
  	}
  	this.appendSection("Free players")
  	var freePlayers = Object.values(Room.players).sort(function(p0,p1) {return sortStrings(p0.alias,p1.alias)})
  	for (var i = 0; i < freePlayers.length; i++) {
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
  	newDiv.appendChild(this.createChevron(true))
  	this.node.appendChild(newDiv)
  }
  appendTable(t) {
  	var newDiv = document.createElement("div")
  	newDiv.className = "Match"
  	var newTable = document.createElement("table")
  	newDiv.appendChild(newTable)
  	var createRow = function(leftText, resource, name) {
  		var tr = document.createElement("tr")
	  	var td00 = document.createElement("td")
	  	td00.appendChild(document.createTextNode(leftText))
	  	var td01 = document.createElement("td")
	  	var dice0 = document.createElement("img")
	  	dice0.setAttribute("src", resource)
	  	dice0.className = "TableDice"
	  	td01.appendChild(dice0)
	  	var td02 = document.createElement("td")
	  	td02.appendChild(document.createTextNode(name))
	  	tr.appendChild(td00)
	  	tr.appendChild(td01)
	  	tr.appendChild(td02)
	  	return tr
  	}
  	
  	var tr0 = createRow(t.dice_num+"🎲", "resources/1a.png", "Igrač0")
  	var tr1 = createRow("", "resources/2b.png", "?")
  	newTable.appendChild(tr0)
  	newTable.appendChild(tr1)
  	newDiv.appendChild(this.createChevron(false))
  	this.node.appendChild(newDiv)
  }
  createChevron(small) {
  	var img = document.createElement("img")
  	img.className = "TableChevronRight"
  	if (small) {
  		img.className += "Small"
  	}
  	img.setAttribute("src", "resources/chevronRight.png")
  	return img
  }
  appendNewGame() {
  	var newDiv = document.createElement("div")
  	newDiv.className = "Player"
  	newDiv.appendChild(document.createTextNode("New game"))
  	newDiv.appendChild(this.createChevron(true))
  	this.node.appendChild(newDiv)
  }
  onRoomInfo(e) {
  	console.log("onRoomInfo event received")
  	this.show()
  }
}

Room.players = []
Room.tables = []