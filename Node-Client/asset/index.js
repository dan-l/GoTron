"use strict";

const PLAYER_RECT_WIDTH = 10;
const PLAYER_RECT_HEIGHT = 10;

const gSocket = io();
const gCanvas = new fabric.Canvas("mainCanvas");

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

function getValidatedObject(msg, expectedProps) {
  // As a reminder, we don't bother with security at all in this project.
  let validatedObject;
  try {
    validatedObject = JSON.parse(msg);
  } catch (e) {
    // TODO: Signal failure
    return null;
  }

  for (let expectedProp of expectedProps) {
    if (!(expectedProp in validatedObject)) {
      return null;
    }
  }

  return validatedObject;
}
