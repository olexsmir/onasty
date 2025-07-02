module Data.Me exposing (Me, decode)

import Iso8601
import Json.Decode as Decode exposing (Decoder)
import Time exposing (Posix)


type alias Me =
    { email : String
    , createdAt : Posix
    }


decode : Decoder Me
decode =
    Decode.map2 Me
        (Decode.field "email" Decode.string)
        (Decode.field "created_at" Iso8601.decoder)
