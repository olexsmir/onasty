module Data.Me exposing (Me, decode)

import Json.Decode as Decode exposing (Decoder)


type alias Me =
    { email : String
    , createdAt : String -- TODO: upgrade to elm/time
    }


decode : Decoder Me
decode =
    Decode.map2 Me
        (Decode.field "email" Decode.string)
        (Decode.field "created_at" Decode.string)
