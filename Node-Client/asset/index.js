"use strict";

const PLAYER_RECT_WIDTH = 50;
const PLAYER_RECT_HEIGHT = 50;
const Direction = {
  UP: "U",
  DOWN: "D",
  LEFT: "L",
  RIGHT: "R",
};

const W = 87;
const A = 65;
const S = 83;
const D = 68;

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

// Whether the game has already ended *for the local player* because we died or
// we won.
var gGameEnded = false;

// Keep track of current direction so we don't send redundant emits.
var curDirection = 0;

function handleKeyPress(event) {
  if (event.keyCode === curDirection) return;

  switch (event.keyCode) {
    case W:
      if (curDirection === S) break;
      curDirection = W;
      gSocket.emit("playerMove", {"direction": Direction.UP});
      break;
    case A:
      if (curDirection === D) break;
      curDirection = A;
      gSocket.emit("playerMove", {"direction": Direction.LEFT});
      break;
    case S:
      if (curDirection === W) break;
      curDirection = S;
      gSocket.emit("playerMove", {"direction": Direction.DOWN});
      break;
    case D:
      if (curDirection === A) break;
      curDirection = D;
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

function hideIntroScreen() {
  let introElem = document.getElementById("intro");
  if (!introElem) {
    throw new Error("'intro' element somehow not present");
  }

  introElem.style.display = "none";
}

/**
 * Renders to the canvas a representation of the given game state.
 *
 * @param {String[][]} state
 *        A "board" object as defined in node.go.
 */
function handleGameStateUpdate(state) {
  console.log('onGameStateUpdate')
  if (!(state instanceof Array)) {
    throw new Error("Passed game state that isn't an array");
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
 * Starts the game when we are paired with enough players.
 */
function startGame(info, direction) {
  curDirection = getDirectionCode(direction);
  window.onkeydown = handleKeyPress;
  hideIntroScreen();
  document.getElementById('stats').innerHTML = '<h3 style="color:black">Player : ' + info + '</h3>';
}

/**
 * Returns keyboard event code for the given direction.
 */
function getDirectionCode(direction) {
     if (direction === Direction.UP) return W;
     if (direction === Direction.DOWN) return S;
     if (direction === Direction.RIGHT) return D;
     if (direction === Direction.LEFT) return A;
}

/**
 * Removes ability to control character.
 */
function onPlayerDeath() {
  if (gGameEnded) {
    return;
  }

  console.log('onPlayerDeath')
  gGameEnded = true;
  window.onkeydown = null;
  document.getElementById("deadMsg").style.display = "inline";
}

/**
 * Player won the game.
 */
function onPlayerVictory() {
  if (gGameEnded) {
    return;
  }
  console.log('onPlayerVictory')
  gGameEnded = true;
  window.onkeydown = null;
  document.getElementById("winMsg").style.display = "inline";
}

function main() {
  console.log('main')
  // Register handlers.
  gSocket.on("startGame", startGame);
  gSocket.on("gameStateUpdate", handleGameStateUpdate);
  gSocket.on("playerDead", onPlayerDeath);
  gSocket.on("playerVictory", onPlayerVictory);
}

main();
