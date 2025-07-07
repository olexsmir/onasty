module Data.Credentials exposing (Credentials, decode)

import Json.Decode as Decode exposing (Decoder)


type alias Credentials =
    { accessToken : String
    , refreshToken : String
    }


decode : Decoder Credentials
decode =
    Decode.map2 Credentials
        (Decode.field "access_token" Decode.string)
        (Decode.field "refresh_token" Decode.string)
