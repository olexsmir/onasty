module Shared.Msg exposing (Msg(..))

import Api
import Data.Credentials exposing (Credentials)
import Time


type Msg
    = GotZone Time.Zone
      -- Auth
    | Logout
    | SignedIn Credentials
      -- Session
    | CheckTokenExpiration Time.Posix
    | TriggerTokenRefresh
    | ApiRefreshTokensResponded (Result Api.Error Credentials)
