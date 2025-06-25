port module Ports exposing (sendToClipboard, sendToLocalStorage)

import Json.Encode


port sendToLocalStorage : { key : String, value : Json.Encode.Value } -> Cmd msg


port sendToClipboard : String -> Cmd msg
