import axios from "axios";
import { backend } from "./base.js";

function getUser(token, cb, errorCb) {
  axios
    .get(`${backend}/user`, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
};

function updateUser(token, user, cb) {
  axios
    .put(`${backend}/user`, user, {
      mode: "no-cors",
      credentials: "same-origin",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    .then((r) => cb(r.data));
};

export { getUser, updateUser };
