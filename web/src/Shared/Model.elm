module Shared.Model exposing (Model)

import Auth.User
import Time


type alias Model =
    { user : Auth.User.SignInStatus
    , timeZone : Time.Zone
    , appURL : String
    }
