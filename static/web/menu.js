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
      item.className = "MenuItem"
      item.onclick = (function(item) {
        return function() {
          app.onMenuItemClicked(item)
        }
      })(this.items[i])
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
