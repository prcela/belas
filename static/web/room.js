class Room {
  constructor(node) {
  	this.node = node
  	document.roomInfoListeners.push(this)
  	wsAPI.roomInfo()
  }
  show() {
  	this.node.innerHTML = "room"
  }
  onRoomInfo(e) {
  	console.log("onRoomInfo event received")
  }
}
