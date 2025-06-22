module Data.Note exposing (CreateResponse, decodeCreateResponse)

import Json.Decode as D exposing (Decoder)


type alias CreateResponse =
    { slug : String }


decodeCreateResponse : Decoder CreateResponse
decodeCreateResponse =
    D.map CreateResponse
        (D.field "slug" D.string)
