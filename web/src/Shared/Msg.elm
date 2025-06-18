module Shared.Msg exposing (Msg(..))

import Time


type Msg
    = GotZone Time.Zone
      -- User
    | Logout
