// src/auth.js

import { onMount, setContext, getContext } from "svelte";
import { writable } from "svelte/store";

//const isLoading = writable(true);
const GAME_CLIENT = {};

function createClient(renderer, characterCreator, muxStore) {
  let ws;
  let messageHandlers = new Map();
  let wsurl = "";

  let mux = muxStore;

  let activeRoom = {};
  let currentCharacter = {};


  messageHandlers["enterRoom"] = (msg) => {
    activeRoom = msg.room;
    renderer(msg.message);

    if (mux) {
      mux.setExits(activeRoom.exits);

      //TODO: set default?
      if (activeRoom.meta != undefined && activeRoom.meta.background != undefined){
        mux.setBackground(activeRoom.meta.background)
      }

      if (activeRoom.actions != undefined) {
        mux.setActions(activeRoom.actions);
      } else {
        mux.setActions([]);
      }
    }
  };

  messageHandlers["createCharacter"] = (msg) => {
    renderer(msg.message);

    if (characterCreator) {
      characterCreator();
      //TODO: send select character
    }
  };
  messageHandlers["characterSelected"] = (msg) => {
    currentCharacter = msg.character;
    renderer(msg.message);
  };

  const setWSClient = async (wscl) => {
    ws = wscl;
    wsurl = ws.url;

    updateClient(ws);
  };

  const updateClient = (ws) => {
    ws.addEventListener("message", function (e) {
      var msg = JSON.parse(e.data);

      if (messageHandlers[msg.type]) {
        messageHandlers[msg.type](msg);
      } else {
        let message = msg.message;

        if (message === "" || message === "\n") {
          console.log("RECEIVED EMPTY MESSAGE")
          return;
        }

        if (msg.username) {
          message = msg.username + ":  " + msg.message;
        }
        renderer(message);
      }
    });

    ws.addEventListener("close", function (e) {
      renderer("Connection Closed.");
    });
  };

  const onInput = async (data) => {
    const msg = await handleInput(data);
    sendMessage(msg);
    //renderer(msg);
  };

  const sendMessage = (msg) => {
    if (!ws) return;

    if (
      ws.readyState == WebSocket.CLOSING ||
      ws.readyState == WebSocket.CLOSED
    ) {
      ws = new WebSocket(wsurl);
      updateClient(ws);
      renderer("reconnecting ...\n");
    }

    ws.send(
      JSON.stringify({
        message: msg,
        type: "message",
      })
    );
  };

  const renderRoom = async (room) => {
    renderer(room.description);
  };

  const handleInput = async (data) => {
    return `${data}`;
  };

  const client = {
    onInput,
    setWSClient,
    sendMessage,
  };

  // setInterval(function () {
  //   renderer("\n<The lights in front of you are flickering>")
  // }, 5000);

  setContext(GAME_CLIENT, client);
  return client;
}

function getClient() {
  return getContext(GAME_CLIENT);
}

export { createClient, getClient };
