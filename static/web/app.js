class App {
  constructor(node) {
    this.node = node
    this.menu = new Menu(["Single player","Multiplayer","Leaderboard","Rules","About"])
    this.onMenuItemClicked = function(index) {
      console.log("klik ",index)
    }
  }
  show() {
    this.menu.show(this.node)
  }
}

class Menu {
  constructor(items) {
    this.items = items
  }
  show(node) {
    node.innerHTML = ""
    var tempList = document.getElementById("tList")
    var clonList = tempList.content.cloneNode(true)
    node.appendChild(clonList)
    for (var i = 0; i < this.items.length; i++) {
      var item = document.createElement('div')
      item.onclick = (function(idx) {
        return function() {
          app.onMenuItemClicked(idx)
        }
      })(i)
      item.addEventListener("click", function(i){
        return function() {
          app.onMenuItemClicked(i)
        }
      })
      item.textContent = this.items[i]
      node.querySelector("#List").appendChild(item)
    }
  }
}

var app = new App(document.getElementById("app_container"))
app.show()