"use strict";

const PLAYER_RECT_WIDTH = 10;
const PLAYER_RECT_HEIGHT = 10;
const Direction = {
  UP: "U",
  DOWN: "D",
  LEFT: "L",
  RIGHT: "R",
};

const gSocket = io();
const gCanvas = new fabric.Canvas("mainCanvas");

// TODO: Remove this at some point. It's here just so we know when the HTTP
//       and JS layers work.
let gPlayerRect = new fabric.Rect({
  left: 100,
  top: 100,
  fill: "red",
  width: PLAYER_RECT_WIDTH,
  height: PLAYER_RECT_HEIGHT,
});
gCanvas.add(gPlayerRect);

// TODO: Get the ID from user input.=
let gUserID = "TEMP USER ID";

window.onkeydown = function(event) {
  switch (event.key) {
    case "w":
      gSocket.emit("playerMove", {"id": gUserID, "direction": Direction.UP});
      break;
    case "a":
      gSocket.emit("playerMove", {"id": gUserID, "direction": Direction.LEFT});
      break;
    case "s":
      gSocket.emit("playerMove", {"id": gUserID, "direction": Direction.DOWN});
      break;
    case "d":
      gSocket.emit("playerMove", {"id": gUserID, "direction": Direction.RIGHT});
      break;
    default:
      return;
  }
};

/**
 * Checks that the given pbject has all properties it is expected to have.
 *
 * @argument {Object} obj
 *           The object to validate.
 * @argument {String[]} expectedProps
 *           Properties the object should have.
 * @returns {Boolean}
 *          true, if it had all the expected properties. false otherwise.
 */
function objContainsProps(obj, expectedProps) {
  for (let expectedProp of expectedProps) {
    if (!(expectedProp in obj)) {
      return false;
    }
  }

  return true;
}

// TODO: This exists as an interim step to getting full game state transmitted
//       back to the JS layer, and should be removed later.
function onPlayerMoveEcho(move) {
  if (!objContainsProps(move, ["id", "direction"])) {
    throw new Error("Move obj does not contain all expected props")
  }

  if (move.id != gUserID) {
    throw new Error("Should not happen in the current impl");
  }

  switch (move.direction) {
    case Direction.UP:
      gPlayerRect.setTop(gPlayerRect.getTop() - gPlayerRect.getHeight());
      break;
    case Direction.LEFT:
      gPlayerRect.setLeft(gPlayerRect.getLeft() - gPlayerRect.getHeight());
      break;
    case Direction.DOWN:
      gPlayerRect.setTop(gPlayerRect.getTop() + gPlayerRect.getHeight());
      break;
    case Direction.RIGHT:
      gPlayerRect.setLeft(gPlayerRect.getLeft() + gPlayerRect.getWidth());
      break;
    default:
      return;
  }

  gCanvas.renderAll();
}
gSocket.on("playerMoveEcho", onPlayerMoveEcho);
