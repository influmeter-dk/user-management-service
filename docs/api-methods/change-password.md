# Change Password

Change the password of the user to the new one after checking that all credentials are correct.

* **URL**

  /v1/change-password

* **Method:**

  `POST`

* **URL Params:**

  * None

* **Data Params:**

  * **Type:** application/json
  * **Content:** `{ "email": "<user mail>", "password": "<user password>", "newPassword": "<new user password>", "newPasswordRepeat": "<new user password>" }`

* **Success Response:**

  * **Code:** 200
  * **Content:** `{ "success": true }`

* **Error Response:**

  * **Code:** 400 Bad Request
  * **Content:** `{ "error": "<error message>" }`
  * **Typical Reason:** Data format (json body of the Post request) wrong, e.g. missing email/password, new password does not match repeated password.

  * **Code:** 401 Unauthorized
  * **Content:** `{ "error": "<error message>" }`
  * **Typical Reason:** Email or password wrong or doesn't belong to any registered participant.

  * **Code:** 500 Internal Server Error
  * **Content:** `{ "error": "<error message>" }`
  * **Typical Reason:** An unexpected error happened during servicing the request.

* **Sample Call:**

  ```go
     creds := &userCredentials{
      Email:             "your@email.com",  // `json:"email"`
      Password:          "yourpassword",    // `json:"password"`
      NewPassword:       "yournewpassword", // `json:"newPassword"`
      NewPasswordRepeat: "yournewpassword", // `json:"newPasswordRepeat"`
    }
    payload, err := json.Marshal(creds)
    res, err := http.Post(user-service-addr + "/v1/user/change-password", "application/json", bytes.NewBuffer(payload))
    defer resp.Body.Close()
  ```

* **Notes:**

  None