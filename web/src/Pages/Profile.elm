module Pages.Profile exposing (Model, Msg, ViewVariant, page)

import Api
import Api.Profile
import Auth
import Components.Box
import Components.Form
import Components.Utils
import Data.Me exposing (Me)
import Effect exposing (Effect)
import Html as H exposing (Html)
import Html.Attributes as A
import Html.Events
import Layouts
import Page exposing (Page)
import Route exposing (Route)
import Shared
import Time.Format
import Validators
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
    , password : { current : String, new : String, confirm : String }
    , email : { current : String, new : String }
    , apiError : Maybe Api.Error
    }


init : Shared.Model -> () -> ( Model, Effect Msg )
init _ () =
    ( { view = Overview
      , me = Api.Loading
      , password = { current = "", new = "", confirm = "" }
      , email = { current = "", new = "" }
      , apiError = Nothing
      }
    , Api.Profile.me { onResponse = ApiMeResponded }
    )



-- UPDATE


type ViewVariant
    = Overview
    | Password
    | Email


type Field
    = PasswordCurrent
    | PasswordNew
    | PasswordConfirm
    | EmailNew


type Msg
    = UserChangedView ViewVariant
    | UserClickedSubmit
    | UserChangedField Field String
    | ApiMeResponded (Result Api.Error Me)
    | ApiChangePasswordResponsed (Result Api.Error ())
    | ApiRequestEmailChangeResponsed (Result Api.Error ())


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        UserChangedView variant ->
            ( { model | view = variant }, Effect.none )

        UserChangedField PasswordCurrent value ->
            ( { model | password = { current = value, new = model.password.new, confirm = model.password.confirm } }, Effect.none )

        UserChangedField PasswordNew value ->
            ( { model | password = { current = model.password.current, new = value, confirm = model.password.confirm } }, Effect.none )

        UserChangedField PasswordConfirm value ->
            ( { model | password = { current = model.password.current, new = model.password.new, confirm = value } }, Effect.none )

        UserChangedField EmailNew value ->
            ( { model | email = { current = model.email.current, new = value } }, Effect.none )

        UserClickedSubmit ->
            case model.view of
                Password ->
                    ( model
                    , Api.Profile.changePassword
                        { onResponse = ApiChangePasswordResponsed
                        , currentPassword = model.password.current
                        , newPassword = model.password.new
                        }
                    )

                Email ->
                    -- TODO: prompt user to look in inbox
                    ( model
                    , Api.Profile.requestEmailChange
                        { onResponse = ApiRequestEmailChangeResponsed
                        , newEmail = model.email.current
                        }
                    )

                _ ->
                    ( model, Effect.none )

        ApiMeResponded (Ok userData) ->
            ( { model | me = Api.Success userData }, Effect.none )

        ApiMeResponded (Err error) ->
            ( { model | me = Api.Failure error }, Effect.none )

        ApiChangePasswordResponsed (Ok ()) ->
            ( model, Effect.none )

        ApiChangePasswordResponsed (Err err) ->
            ( { model | apiError = Just err }, Effect.none )

        ApiRequestEmailChangeResponsed (Ok ()) ->
            ( model, Effect.none )

        ApiRequestEmailChangeResponsed (Err err) ->
            ( { model | apiError = Just err }, Effect.none )


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
                    [ Components.Utils.viewMaybe model.apiError (\e -> Components.Box.error (Api.errorMessage e))
                    , H.h1 [ A.class "text-2xl font-bold text-gray-900" ] [ H.text "Account Settings" ]
                    , H.p [ A.class "text-gray-600 mt-2" ] [ H.text "Manage your account preferences and security settings" ]
                    ]
                , H.div [ A.class "flex" ]
                    [ viewNavigationSidebar model
                    , H.div [ A.class "flex-1 p-6" ]
                        [ case model.me of
                            Api.Success me ->
                                case model.view of
                                    Overview ->
                                        viewOverview shared me

                                    Password ->
                                        viewPassword model.password (isFormDisabled model)

                                    Email ->
                                        viewEmail me model.email (isFormDisabled model)

                            Api.Loading ->
                                H.text "Loading..."

                            Api.Failure err ->
                                H.text ("ERROR: " ++ Api.errorMessage err)
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
            ]
        ]


viewOverview : Shared.Model -> Me -> Html Msg
viewOverview shared me =
    let
        infoBox title text =
            H.div [ A.class "bg-gray-50 rounded-lg p-4" ]
                [ H.div [ A.class "flex items-center gap-3 mb-2" ]
                    [ H.h3 [ A.class "font-medium text-gray-900" ] [ H.text title ] ]
                , H.p [ A.class "text-gray-700" ] [ H.text text ]
                ]
    in
    viewWrapper
        { title = "Account Overview"
        , body =
            H.div [ A.class "grid grid-cols-1 md:grid-cols-2 gap-6" ]
                [ infoBox "Email Address" me.email
                , infoBox "Member Since" (Time.Format.toString shared.timeZone me.createdAt)
                , infoBox "Last Login" (Time.Format.toString shared.timeZone me.lastLoginAt)
                , infoBox "Total Notes Created" (String.fromInt me.notesCreated)
                ]
        }


viewPassword : { current : String, new : String, confirm : String } -> Html Msg
viewPassword password =
    let
        input : { label : String, field : Field, value : String, error : Maybe String } -> Html Msg
        input { label, field, value, error } =
            Components.Form.input
                { label = label
                , id = label
                , field = field
                , onInput = UserChangedField field
                , placeholder = ""
                , value = value
                , required = True
                , type_ = "password"
                , style = Components.Form.Simple
                , error = error
                }
    in
    viewWrapper
        { title = "Change Password"
        , body =
            H.form
                [ A.class "space-y-4 max-w-md"
                , Html.Events.onSubmit UserClickedSubmit
                ]
                [ input { label = "Current Password", field = PasswordCurrent, value = password.current, error = Nothing }
                , input { label = "New Password", field = PasswordNew, value = password.new, error = Validators.password password.new }
                , input { label = "Confirm New Password", field = PasswordConfirm, value = password.confirm, error = Validators.passwords password.new password.confirm }
                , Components.Form.submitButton
                    { disabled = isButtonDisabled
                    , text = "Change Password"
                    , style = Components.Form.Primary False
                    , class = ""
                    }
                ]
        }


viewEmail : Me -> { current : String, new : String } -> Html Msg
viewEmail me email =
    viewWrapper
        { title = "Change Email Address"
        , body =
            H.form
                [ A.class "space-y-4 max-w-md"
                , Html.Events.onSubmit UserClickedSubmit
                ]
                [ H.div [ A.class "mb-6 p-4 bg-blue-50 border border-blue-200 rounded-md" ]
                    [ H.h3 [ A.class "font-medium mb-1" ] [ H.text "Note:" ]
                    , H.p [] [ H.text "A confirmation email will be sent to your new email address. You must confirm the change by clicking the link in that email." ]
                    , H.p [ A.class "mt-2 text-blue-800 text-sm" ]
                        [ H.span [ A.class "font-medium" ] [ H.text ("Current email: " ++ me.email) ]
                        ]
                    ]
                , Components.Form.input
                    { style = Components.Form.Simple
                    , id = "new-email"
                    , type_ = "email"
                    , field = EmailNew
                    , label = "New Email Address"
                    , value = email.current
                    , placeholder = "Enter your new email address"
                    , onInput = UserChangedField EmailNew
                    , error = Validators.email email
                    , required = True
                    }
                , Components.Form.submitButton
                    { disabled = isButtonDisabled
                    , text = "Update Email"
                    , style = Components.Form.Primary False
                    , class = ""
                    }
                ]
        }


viewWrapper : { title : String, body : Html Msg } -> Html Msg
viewWrapper { title, body } =
    H.div [ A.class "space-y-6" ]
        [ H.div []
            [ H.h2 [ A.class "text-lg font-semibold text-gray-900 mb-4" ] [ H.text title ]
            , body
            ]
        ]
