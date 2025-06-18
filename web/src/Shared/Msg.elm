module Shared.Msg exposing (Msg(..))

import Data.Credentials exposing (Credentials)
import Http
import Time


type Msg
    = GotZone Time.Zone
      -- Auth
    | Logout
    | SignedIn Credentials
      -- Session
    | CheckTokenExpiration Time.Posix
    | TriggerTokenRefresh
    | ApiRefreshTokensResponded (Result Http.Error Credentials)
