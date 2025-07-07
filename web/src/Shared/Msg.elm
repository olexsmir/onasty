module Shared.Msg exposing (Msg(..))

import Api
import Data.Credentials exposing (Credentials)
import Time exposing (Posix, Zone)


type Msg
    = GotZone Zone
      -- Auth
    | Logout
    | SignedIn Credentials
      -- Session
    | CheckTokenExpiration Posix
    | TriggerTokenRefresh
    | ApiRefreshTokensResponded (Result Api.Error Credentials)
