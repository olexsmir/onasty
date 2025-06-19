module Shared.Model exposing (Model)

import Auth.User
import Data.Credentials exposing (Credentials)
import Time


type alias Model =
    { user : Auth.User.SignInStatus
    , timeZone : Time.Zone
    , isRefreshingTokens : Bool
    }
