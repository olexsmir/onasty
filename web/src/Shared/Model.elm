module Shared.Model exposing (Model)

import Data.Credentials exposing (Credentials)
import Time


type alias Model =
    { credentials : Maybe Credentials
    , timeZone : Time.Zone
    }
