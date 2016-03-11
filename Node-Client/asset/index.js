"use strict";

var gSocket = io();

var gCanvas = new fabric.Canvas("mainCanvas");
var gPlayerRect = new fabric.Rect({
  left: 100,
  top: 100,
  fill: "red",
  width: 10,
  height: 10
});
gCanvas.add(gPlayerRect);

window.onkeydown = function(event) {
  switch (event.key) {
    case "w":
      gPlayerRect.setTop(Math.max(gPlayerRect.getTop() - gPlayerRect.getHeight(), 0));
      gSocket.emit("playerMove", "U");
      break;
    case "a":
      gPlayerRect.setLeft(Math.max(gPlayerRect.getLeft() - gPlayerRect.getHeight(), 0));
      gSocket.emit("playerMove", "L");
      break;
    case "s":
      gPlayerRect.setTop(Math.min(gPlayerRect.getTop() + gPlayerRect.getHeight(), 200));
      gSocket.emit("playerMove", "D");
      break;
    case "d":
      gPlayerRect.setLeft(Math.min(gPlayerRect.getLeft() + gPlayerRect.getWidth(), 200));
      gSocket.emit("playerMove", "R");
      break;
    default:
      return;
  }

  gCanvas.renderAll();
};
