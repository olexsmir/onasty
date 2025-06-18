module Shared.Msg exposing (Msg(..))

import Data.Credentials exposing (Credentials)
import Time


type Msg
    = GotZone Time.Zone
      -- User
    | Logout
    | SignedIn Credentials
