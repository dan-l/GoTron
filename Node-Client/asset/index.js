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
let gUserCodeToIDMap;

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
function handleGameStateUpdate(state) {
  if (!(state instanceof Array)) {
    throw new Error("Passed game state that isn't an array");
  }

  if (!gUserCodeToIDMap) {
    throw new Error("User code to ID map not init")
  }

  for (let y = 0; y < state.length; y ++) {
    let row = state[y];
    if (!(row instanceof Array)) {
      throw new Error("Passed row that isn't an array");
    }

    for (let x = 0; x < row.length; x++) {
      if (!(state[y][x] in gUserCodeToIDMap)) {
        continue;
      }

      // TODO: Handle additional players.
      if (state[y][x] == "p1") {
        gPlayerRect.setTop(y * gPlayerRect.getHeight());
        gPlayerRect.setLeft(x * gPlayerRect.getWidth());
      }
    }
  }

  gCanvas.renderAll();
}

// TODO: Document.
function handleInitialConfig(initialConfig) {
  if (!objContainsProps(initialConfig, ["Players"])) {
    throw new Error("Got invalid initial config");
  }

  // TODO: At bare minimum, |initialConfig.players| should be an object with p1
  //       *and* p2 defined, since the minimum number of players is 2. We should
  //       catch that.
  if (!objContainsProps(initialConfig.Players, ["p1"])) {
    throw new Error("Got invalid initial config players");
  }

  gUserCodeToIDMap = initialConfig.Players;
}

// TODO: Document.
function sendPlayerInfo() {
  gSocket.emit("playerInfo", {"id": gUserID});
}

function main() {
  gCanvas.add(gPlayerRect);

  // Register handlers.
  gSocket.on("initialConfig", handleInitialConfig);
  gSocket.on("gameStateUpdate", handleGameStateUpdate);

  sendPlayerInfo();

  window.onkeydown = handleKeyPress;
}

main();
