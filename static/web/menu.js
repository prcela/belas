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
      item.onclick = (function(menu,item) {
        return function() {
          menu.onMenuItemClicked(item)
        }
      })(this,this.items[i])
      item.textContent = this.items[i]
      this.node.querySelector("#List").appendChild(item)
    }
  }

  onMenuItemClicked(item) {
    console.log("klik ",item)
    switch (item) {
      case "Multiplayer": 
        var room = new RoomViewController(document.createElement("div"))
        this.navigationController.push(room)
        break
      case "Rules": 
        var rules = new RulesViewController(document.createElement("div"))
        this.navigationController.push(rules)
        break
      case "About": 
        var about = new AboutViewController(document.createElement("div"))
        this.navigationController.push(about)
        break
    }
  }

}

