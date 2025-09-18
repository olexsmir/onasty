port module Ports exposing (confirmRequest, confirmResponse, sendToClipboard, sendToLocalStorage)

import Json.Encode


port sendToLocalStorage : { key : String, value : Json.Encode.Value } -> Cmd msg


port sendToClipboard : String -> Cmd msg


port confirmRequest : String -> Cmd msg


port confirmResponse : (Bool -> msg) -> Sub msg
