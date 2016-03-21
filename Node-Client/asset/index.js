"use strict";

const PLAYER_RECT_WIDTH = 10;
const PLAYER_RECT_HEIGHT = 10;
const Direction = {
  UP: "U",
  DOWN: "D",
  LEFT: "L",
  RIGHT: "R",
};

// Maps player codes constants such as "p1" and "t1" to a colour.
const PLAYER_CODE_TO_COLOUR = {
  "p1": "red",
  "t1": "red",
  "p2": "green",
  "t2": "green",
  "p3": "blue",
  "t3": "blue",
  "p4": "orange",
  "t4": "orange",
  "p5": "brown",
  "t5": "brown",
  "p6": "black",
  "t6": "black",
};

const gSocket = io();
// We use a StaticCanvas since we don't want users to be able to be able to
// perform interactions such as resizing objects.
const gCanvas = new fabric.StaticCanvas("mainCanvas");

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

  // For now, we want to throw away the existing canvas and repaint everything
  // whenever we get an update. All of this is pretty inefficient, but probably
  // serves the requirements of this project well enough.
  gCanvas.dispose();

  for (let y = 0; y < state.length; y ++) {
    let row = state[y];
    if (!(row instanceof Array)) {
      throw new Error("Passed row that isn't an array");
    }

    for (let x = 0; x < row.length; x++) {
      let playerCode = state[y][x];
      if (playerCode.length != 2) {
        continue;
      }

      if (!(playerCode in PLAYER_CODE_TO_COLOUR)) {
        throw new Error("State contains unknown player code: " + playerCode);
      }

      // TODO: Handle additional players.
      let canvasProps = {
        left: x * PLAYER_RECT_WIDTH,
        top: y * PLAYER_RECT_HEIGHT,
        width: PLAYER_RECT_WIDTH,
        height: PLAYER_RECT_HEIGHT,
        fill: PLAYER_CODE_TO_COLOUR[playerCode],
      };
      // If this is a trail, lower the opacity to make it visually obvious.
      if (playerCode.charAt(0) == "t") {
        canvasProps.opacity = 0.5;
      }
      gCanvas.add(new fabric.Rect(canvasProps));
    }
  }
}

/**
 * Sets up variables etc with the given initial config.
 *
 * @param {Object} initialConfig
 *        An InitialConfig object as defined in httpServer.go.
 */
function handleInitialConfig(initialConfig) {
  if (!objContainsProps(initialConfig, ["Players"])) {
    throw new Error("Got invalid initial config");
  }

  // Since the minimum number of players is 2, we can only test that at least
  // p1 and p2 are defined.
  if (!objContainsProps(initialConfig.Players, ["p1", "p2"])) {
    throw new Error("Got invalid initial config players");
  }

  gUserCodeToIDMap = initialConfig.Players;
}

// TODO: Document.
function sendPlayerInfo() {
  gSocket.emit("playerInfo", {"id": gUserID});
}

function main() {
  // Register handlers.
  gSocket.on("initialConfig", handleInitialConfig);
  gSocket.on("gameStateUpdate", handleGameStateUpdate);

  sendPlayerInfo();

  window.onkeydown = handleKeyPress;
}

main();
