import React, { useCallback, useMemo, useState } from 'react'

import classNames from 'classnames'
import cookies from 'js-cookie'
import { Observable, of } from 'rxjs'
import { fromFetch } from 'rxjs/fetch'
import { catchError, switchMap } from 'rxjs/operators'

import { asError } from '@sourcegraph/common'
import {
    useInputValidation,
    ValidationOptions,
    deriveInputClassName,
} from '@sourcegraph/shared/src/util/useInputValidation'
import { Link, Label, Text, LoaderInput, ErrorAlert } from '@sourcegraph/wildcard'

import { LoaderButton } from '../components/LoaderButton'
import { SourcegraphContext } from '../jscontext'
import { ANONYMOUS_USER_ID_KEY, eventLogger, FIRST_SOURCE_URL_KEY, LAST_SOURCE_URL_KEY } from '../tracking/eventLogger'
import { validatePassword, getPasswordRequirements } from '../util/security'

import { PasswordInput, UsernameInput } from './SignInSignUpCommon'
import { SignupEmailField } from './SignupEmailField'

import signInSignUpCommonStyles from './SignInSignUpCommon.module.scss'

export interface SignUpArguments {
    email: string
    username: string
    password: string
    anonymousUserId?: string
    firstSourceUrl?: string
    lastSourceUrl?: string
}

interface SignUpFormProps {
    className?: string

    /** Called to perform the signup on the server. */
    onSignUp: (args: SignUpArguments) => Promise<void>

    buttonLabel?: string
    context: Pick<
        SourcegraphContext,
        'authProviders' | 'sourcegraphDotComMode' | 'authPasswordPolicy' | 'authMinPasswordLength'
    >

    // For use in ExperimentalSignUpPage. Modifies styling and removes terms of service.
    experimental?: boolean
}

const preventDefault = (event: React.FormEvent): void => event.preventDefault()

/**
 * The form for creating an account
 */
export const SignUpForm: React.FunctionComponent<React.PropsWithChildren<SignUpFormProps>> = ({
    onSignUp,
    buttonLabel,
    className,
    context,
    experimental = false,
}) => {
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<Error | null>(null)

    const signUpFieldValidators: Record<'email' | 'username' | 'password', ValidationOptions> = useMemo(
        () => ({
            email: {
                synchronousValidators: [],
                asynchronousValidators: [],
            },
            username: {
                synchronousValidators: [],
                asynchronousValidators: [isUsernameUnique],
            },
            password: {
                synchronousValidators: [password => validatePassword(context, password)],
                asynchronousValidators: [],
            },
        }),
        [context]
    )

    const [emailState, nextEmailFieldChange, emailInputReference] = useInputValidation(signUpFieldValidators.email)

    const [usernameState, nextUsernameFieldChange, usernameInputReference] = useInputValidation(
        signUpFieldValidators.username
    )

    const [passwordState, nextPasswordFieldChange, passwordInputReference] = useInputValidation(
        signUpFieldValidators.password
    )

    const canRegister = emailState.kind === 'VALID' && usernameState.kind === 'VALID' && passwordState.kind === 'VALID'

    const disabled = loading || !canRegister

    const handleSubmit = useCallback(
        (event: React.FormEvent<HTMLFormElement>): void => {
            event.preventDefault()
            if (disabled) {
                return
            }

            setLoading(true)
            onSignUp({
                email: emailState.value,
                username: usernameState.value,
                password: passwordState.value,
                anonymousUserId: cookies.get(ANONYMOUS_USER_ID_KEY),
                firstSourceUrl: cookies.get(FIRST_SOURCE_URL_KEY),
                lastSourceUrl: cookies.get(LAST_SOURCE_URL_KEY),
            }).catch(error => {
                setError(asError(error))
                setLoading(false)
            })
            eventLogger.log('InitiateSignUp')
        },
        [onSignUp, disabled, emailState, usernameState, passwordState]
    )

    return (
        <>
            {error && <ErrorAlert className="mt-4 mb-0" error={error} />}
            {/* Using  <form /> to set 'valid' + 'is-invaild' at the input level */}
            {/* eslint-disable-next-line react/forbid-elements */}
            <form
                className={classNames(
                    !experimental && signInSignUpCommonStyles.signinSignupForm,
                    'test-signup-form',
                    !experimental && 'rounded p-4',
                    'text-left',
                    !experimental && error ? 'mt-3' : 'mt-4',
                    className
                )}
                onSubmit={handleSubmit}
                noValidate={true}
            >
                <SignupEmailField
                    label="Email"
                    loading={loading}
                    nextEmailFieldChange={nextEmailFieldChange}
                    emailState={emailState}
                    emailInputReference={emailInputReference}
                />
                <div className="form-group d-flex flex-column align-content-start">
                    <Label
                        htmlFor="username"
                        className={classNames('align-self-start', {
                            'text-danger font-weight-bold': usernameState.kind === 'INVALID',
                        })}
                    >
                        Username
                    </Label>
                    <LoaderInput
                        className={classNames(deriveInputClassName(usernameState))}
                        loading={usernameState.kind === 'LOADING'}
                    >
                        <UsernameInput
                            className={deriveInputClassName(usernameState)}
                            onChange={nextUsernameFieldChange}
                            value={usernameState.value}
                            required={true}
                            disabled={loading}
                            placeholder=" "
                            inputRef={usernameInputReference}
                            aria-describedby="username-input-invalid-feedback"
                        />
                    </LoaderInput>
                    {usernameState.kind === 'INVALID' && (
                        <small className="invalid-feedback" id="username-input-invalid-feedback" role="alert">
                            {usernameState.reason}
                        </small>
                    )}
                </div>
                <div className="form-group d-flex flex-column align-content-start">
                    <Label
                        htmlFor="password"
                        className={classNames('align-self-start', {
                            'text-danger font-weight-bold': passwordState.kind === 'INVALID',
                        })}
                    >
                        Password
                    </Label>
                    <LoaderInput
                        className={classNames(deriveInputClassName(passwordState))}
                        loading={passwordState.kind === 'LOADING'}
                    >
                        <PasswordInput
                            className={deriveInputClassName(passwordState)}
                            onChange={nextPasswordFieldChange}
                            value={passwordState.value}
                            required={true}
                            disabled={loading}
                            autoComplete="new-password"
                            minLength={context.authMinPasswordLength}
                            placeholder=" "
                            onInvalid={preventDefault}
                            inputRef={passwordInputReference}
                            formNoValidate={true}
                            aria-describedby="password-input-invalid-feedback password-requirements"
                        />
                    </LoaderInput>
                    {passwordState.kind === 'INVALID' && (
                        <small className="invalid-feedback" id="password-input-invalid-feedback" role="alert">
                            {passwordState.reason}
                        </small>
                    )}
                    <small className="form-help text-muted" id="password-requirements">
                        {getPasswordRequirements(context)}
                    </small>
                </div>
                <div className="form-group mb-0">
                    <LoaderButton
                        loading={loading}
                        label={buttonLabel || 'Register'}
                        type="submit"
                        disabled={disabled}
                        variant="primary"
                        display="block"
                    />
                </div>
            </form>
        </>
    )
}

// Asynchronous Validators

function isUsernameUnique(username: string): Observable<string | undefined> {
    return fromFetch(`/-/check-username-taken/${username}`).pipe(
        switchMap(response => {
            switch (response.status) {
                case 200:
                    return of('Username is already taken.')
                case 404:
                    // Username is unique
                    return of(undefined)

                default:
                    return of('Unknown error validating username')
            }
        }),
        catchError(() => of('Unknown error validating username'))
    )
}
