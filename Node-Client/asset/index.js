"use strict";

const PLAYER_RECT_WIDTH = 16;
const PLAYER_RECT_HEIGHT = 16;
const Direction = {
  UP: "U",
  DOWN: "D",
  LEFT: "L",
  RIGHT: "R",
};

// Maps player codes constants such as "p1" and "t1" to a colour.
const PLAYER_CODE_TO_COLOUR = {
  "d1": "red",
  "p1": "red",
  "t1": "red",
  "d2": "green",
  "p2": "green",
  "t2": "green",
  "d3": "blue",
  "p3": "blue",
  "t3": "blue",
  "d4": "orange",
  "p4": "orange",
  "t4": "orange",
  "d5": "brown",
  "p5": "brown",
  "t5": "brown",
  "d6": "black",
  "p6": "black",
  "t6": "black",
};

const gSocket = io();
// We use a StaticCanvas since we don't want users to be able to be able to
// perform interactions such as resizing objects.
const gCanvas = new fabric.StaticCanvas("mainCanvas");

let gPlayerID;

function handleKeyPress(event) {
  switch (event.keyCode) {
    case 87: // 'w'
      gSocket.emit("playerMove", {"direction": Direction.UP});
      break;
    case 65: // 'a'
      gSocket.emit("playerMove", {"direction": Direction.LEFT});
      break;
    case 83: // 's'
      gSocket.emit("playerMove", {"direction": Direction.DOWN});
      break;
    case 68: // 'd'
      gSocket.emit("playerMove", {"direction": Direction.RIGHT});
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

/**
 * Renders to the canvas a representation of the given game state.
 *
 * @param {String[][]} state
 *        A "board" object as defined in node.go.
 */
function handleGameStateUpdate(state) {
  if (!(state instanceof Array)) {
    throw new Error("Passed game state that isn't an array");
  }

  if (!gPlayerID) {
    throw new Error("User ID not set")
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
      // If the player is dead, we want to overlay a indicator on top.
      if (playerCode.charAt(0) == "d") {
        gCanvas.add(new fabric.Line([
          canvasProps.left,
          canvasProps.top,
          canvasProps.left + canvasProps.width,
          canvasProps.top + canvasProps.height,
        ], {
          fill: "white",
          stroke: "white",
          strikeWidth: 10,
        }));
      }
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
  if (!objContainsProps(initialConfig, ["LocalID"])) {
    throw new Error("Got invalid initial config");
  }

  gPlayerID = initialConfig.LocalID;
}

/**
 * Starts the game when we are paired with enough players.
 */
function startGame() {
   gSocket.on("gameStateUpdate", handleGameStateUpdate);
   gSocket.on("playerDead", onPlayerDeath)
   window.onkeydown = handleKeyPress;
   document.getElementById('intro').style.display = 'none';
}

/**
 * Removes ability to control character.
 */
function onPlayerDeath() {
    window.onkeydown = null;
    document.getElementById('message').innerHTML = '<h3 style="color:red">You are dead!</h3>'
}

function main() {
  // Register handlers.
  gSocket.on("initialConfig", handleInitialConfig);
  gSocket.on("startGame", startGame);
}

main();
