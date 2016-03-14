"use strict";

var PLAYER_RECT_WIDTH = 10;
var PLAYER_RECT_HEIGHT = 10;
var PLAYER_COLOURS = [
  "red",
  "green",
  "blue",
  "yellow",
  "black",
  "orange",
];

var gSocket = io();
var gCanvas = new fabric.Canvas("mainCanvas");
var gPlayerRects = {};
var gPlayerRect;
var gInitialised = false;

window.onkeydown = function(event) {
  if (!gInitialised) {
    return;
  }

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

gSocket.on("config", function(msg) {
  if (gInitialised) {
    // TODO: Signal failure
    return;
  }

  // As a reminder, we don't bother with security at all in this project.
  var config;
  try {
    config = JSON.parse(msg);
  } catch (e) {
    // TODO: Signal failure
    return;
  }

  var EXPECTED_CONFIG_PROPS = [
    "players",
    "selfID",
  ];
  for (var i = 0; i < EXPECTED_CONFIG_PROPS.length; i++) {
    var expectedProp = EXPECTED_CONFIG_PROPS[i];
    if (!(expectedProp in config)) {
      console.log("missing prop: " + expectedProp);
      // TODO: Signal failure
      return;
    }
  }

  if (config.players.length > PLAYER_COLOURS.length) {
    console.log("player count too high");
    // TODO: Signal failure
    return;
  }

  for (var i = 0; i < config.players.length; i++) {
    var rect = new fabric.Rect({
      left: i * 20,
      top: i * 20,
      fill: PLAYER_COLOURS[i],
      width: PLAYER_RECT_WIDTH,
      height: PLAYER_RECT_HEIGHT,
    });
    gPlayerRects[config.players[i]] = rect;
    gPlayerRect = gPlayerRects[config.selfID];
    gCanvas.add(rect);
  }

  gCanvas.renderAll();
  gInitialised = true;
});
