port module Ports exposing (sendToLocalStorage)

import Json.Encode


port sendToLocalStorage : { key : String, value : Json.Encode.Value } -> Cmd msg
