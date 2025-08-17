module Pages.Profile exposing (Model, Msg, ViewVariant, page)

import Api
import Api.Me
import Auth
import Components.Form
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
        UserChangedView variant ->
            ( { model | view = variant }, Effect.none )

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
        -- FIXME: feels like there's a lot of redundant classes here
        [ H.div [ A.class "w-full p-6 max-w-4xl mx-auto" ]
            [ H.div [ A.class "bg-white rounded-lg border border-gray-200 shadow-sm" ]
                [ H.div [ A.class "p-6 pb-4 border-b border-gray-200" ]
                    [ H.h1 [ A.class "text-2xl font-bold text-gray-900" ] [ H.text "Account Settings" ]
                    , H.p [ A.class "text-gray-600 mt-2" ] [ H.text "Manage your account preferences and security settings" ]
                    ]
                , H.div [ A.class "flex" ]
                    [ viewNavigationSidebar model
                    , H.div [ A.class "flex-1 p-6" ]
                        [ case model.view of
                            Overview ->
                                viewOverview shared model.me

                            Password ->
                                H.text "Password View"

                            Email ->
                                H.text "Email View"

                            DeleteAccount ->
                                H.text "Delete Account View"
                        ]
                    ]
                ]
            ]
        ]
    }


viewNavigationSidebar : Model -> Html Msg
viewNavigationSidebar model =
    let
        button variant text =
            -- TODO: add icons to buttons
            Components.Form.button
                { text = text
                , onClick = UserChangedView variant
                , disabled = model.view == variant
                , style = Components.Form.PrimaryReverse (model.view == variant)
                }
    in
    H.div [ A.class "w-64 border-r border-gray-200 p-6" ]
        [ H.nav [ A.class "[&>*]:w-full space-y-2" ]
            [ button Overview "Overview"
            , button Password "Password"
            , button Email "Email"
            , button DeleteAccount "Delete Account"
            ]
        ]


viewOverview : Shared.Model -> Api.Response Me -> Html Msg
viewOverview shared userResponse =
    let
        infoBox title text =
            H.div [ A.class "bg-gray-50 rounded-lg p-4" ]
                [ H.div [ A.class "flex items-center gap-3 mb-2" ]
                    [ H.h3 [ A.class "font-medium text-gray-900" ] [ H.text title ] ]
                , H.p [ A.class "text-gray-700" ] [ H.text text ]
                ]
    in
    genericResponseView userResponse <|
        \user ->
            H.div [ A.class "space-y-6" ]
                [ H.div []
                    [ H.h2 [ A.class "text-lg font-semibold text-gray-900 mb-4" ]
                        [ H.text "Account Overview" ]
                    , H.div [ A.class "grid grid-cols-1 md:grid-cols-2 gap-6" ]
                        [ infoBox "Email Address" user.email
                        , infoBox "Member Since" (Time.Format.toString shared.timeZone user.createdAt)
                        , infoBox "Last Login" (Time.Format.toString shared.timeZone user.lastLoginAt)
                        , infoBox "Total Notes Created" (String.fromInt user.notesCreated)
                        ]
                    ]
                ]


genericResponseView : Api.Response a -> (a -> Html Msg) -> Html Msg
genericResponseView apiResp userView =
    case apiResp of
        Api.Success data ->
            userView data

        Api.Loading ->
            H.text "Loading..."

        Api.Failure err ->
            H.text ("ERROR: " ++ Api.errorMessage err)
