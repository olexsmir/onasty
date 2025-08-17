module Pages.Profile exposing (Model, Msg, ViewVariant, page)

import Api
import Api.Me
import Auth
import Data.Me exposing (Me)
import Effect exposing (Effect)
import Html as H exposing (Html)
import Html.Attributes as A
import Layouts
import Page exposing (Page)
import Route exposing (Route)
import Shared
import Time.Format
import View exposing (View)


page : Auth.User -> Shared.Model -> Route () -> Page Model Msg
page _ shared _ =
    Page.new
        { init = init shared
        , update = update
        , subscriptions = subscriptions
        , view = view shared
        }
        |> Page.withLayout (\_ -> Layouts.Header {})



-- INIT


type alias Model =
    { view : ViewVariant
    , me : Api.Response Me
    }


init : Shared.Model -> () -> ( Model, Effect Msg )
init _ () =
    ( { view = Overview
      , me = Api.Loading
      }
    , Api.Me.get { onResponse = ApiMeResponded }
    )



-- UPDATE


type ViewVariant
    = Overview
    | Password
    | Email
    | DeleteAccount


type Msg
    = UserChangedView ViewVariant
    | ApiMeResponded (Result Api.Error Me)


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        UserChangedView _ ->
            ( model, Effect.none )

        ApiMeResponded (Ok userData) ->
            ( { model | me = Api.Success userData }, Effect.none )

        ApiMeResponded (Err error) ->
            ( { model | me = Api.Failure error }, Effect.none )


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Shared.Model -> Model -> View Msg
view shared model =
    { title = "Profile"
    , body =
        [ case model.view of
            Overview ->
                viewProfileOverview shared model.me

            Password ->
                H.text "Password View"

            Email ->
                H.text "Email View"

            DeleteAccount ->
                H.text "Delete Account View"
        ]
    }


viewProfileOverview : Shared.Model -> Api.Response Me -> Html Msg
viewProfileOverview shared userResponse =
    case userResponse of
        Api.Success user ->
            H.div []
                [ H.h1 [] [ H.text "Profile Overview" ]
                , H.p [] [ H.text ("Created at: " ++ Time.Format.toString shared.timeZone user.createdAt) ]
                , H.p [] [ H.text ("Email: " ++ user.email) ]
                ]

        Api.Loading ->
            H.text "Loading..."

        Api.Failure err ->
            H.text (Api.errorMessage err)
