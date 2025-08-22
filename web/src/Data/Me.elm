module Data.Me exposing (Me, decode)

import Iso8601
import Json.Decode as Decode exposing (Decoder)
import Time exposing (Posix)


type alias Me =
    { email : String
    , createdAt : Posix
    , lastLoginAt : Posix
    , notesCreated : Int
    }


decode : Decoder Me
decode =
    Decode.map4 Me
        (Decode.field "email" Decode.string)
        (Decode.field "created_at" Iso8601.decoder)
        (Decode.field "last_login_at" Iso8601.decoder)
        (Decode.field "notes_created" Decode.int)
