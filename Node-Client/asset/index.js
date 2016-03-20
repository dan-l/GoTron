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
  left: 0,
  top: 0,
  fill: "red",
  width: PLAYER_RECT_WIDTH,
  height: PLAYER_RECT_HEIGHT,
});

// TODO: Get the ID from user input.=
let gUserID = "TEMP USER ID";

function handleKeyPress(event) {
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
 * Checks that the given object has all properties it is expected to have.
 *
 * @argument {Object} obj
 *           The object to validate.
 * @argument {String[]} expectedProps
 *           Properties the object should have.
 * @returns {Boolean}
 *          true, if the given object was non-null and had all the expected
 *          properties. false otherwise.
 */
function objContainsProps(obj, expectedProps) {
  if (!expectedProps) {
    throw new Error("A non-null expectedProps must be provided");
  }

  if (!obj) {
    return false;
  }

  for (let expectedProp of expectedProps) {
    if (!(expectedProp in obj)) {
      return false;
    }
  }

  return true;
}

// TODO: Document.
function onGameStateUpdate(state) {
  if (!(state instanceof Array)) {
    throw new Error("Passed game state that isn't an array");
  }

  for (let y = 0; y < state.length; y ++) {
    let row = state[y];
    if (!(row instanceof Array)) {
      throw new Error("Passed row that isn't an array");
    }

    for (let x = 0; x < row.length; x++) {
      if (state[y][x] == gUserID) {
        gPlayerRect.setTop(y * gPlayerRect.getHeight());
        gPlayerRect.setLeft(x * gPlayerRect.getWidth());
      }
    }
  }

  gCanvas.renderAll();
}

function main() {
  gCanvas.add(gPlayerRect);

  // Register handlers.
  gSocket.on("gameStateUpdate", onGameStateUpdate);
  window.onkeydown = handleKeyPress;
}

main();
