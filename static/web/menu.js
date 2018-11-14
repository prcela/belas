class MenuViewController extends ViewController {
  constructor(node, items) {
    super(node)
    node.className = "MenuViewController"
    this.items = items
  }
  show() {
    super.show()
    this.node.innerHTML = ""
    var tempList = document.getElementById("tList")
    var clonList = tempList.content.cloneNode(true)
    this.node.appendChild(clonList)
    for (var i = 0; i < this.items.length; i++) {
      var item = document.createElement('div')
      item.className = "MenuItem"
      item.onclick = (function(item) {
        return function() {
          this.onMenuItemClicked(item)
        }
      })(this.items[i])
      item.textContent = this.items[i]
      this.node.querySelector("#List").appendChild(item)
    }
  }

  onMenuItemClicked(item) {
      console.log("klik ",item)
      if (item == "Multiplayer") {
        var room = new RoomViewController(this.node.firstElementChild)
        this.navigationController.push(room)
      }
    }
}
