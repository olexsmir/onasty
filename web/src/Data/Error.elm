module Data.Error exposing (Error, decode)

import Json.Decode


type alias Error =
    { message : String }


decode : Json.Decode.Decoder Error
decode =
    Json.Decode.map Error
        (Json.Decode.field "message" Json.Decode.string)
