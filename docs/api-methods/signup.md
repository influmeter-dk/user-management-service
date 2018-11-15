**Signup**
----
  Register a new user account with the given email address and password, if they match the validation criterias (valid email format and password at least 6 characters including letters and numbers).

* **URL**

  /v1/signup

* **Method:**

  `POST`

*  **URL Params**
  * None

* **Data Params**
  **Required:**
  * **Type:** application/json <br />
    **Content:** `{ "email": "<user email>", "password": "<user password>"}`

* **Success Response:**

  * **Code:** 201 <br />
    **Content:** `{ "user_id": "<user id>", "role": "PARTICIPANT" | "RESEARCHER" | "ADMIN" }`

* **Error Response:**

  * **Code:** 400 Bad request <br />
    **Content:** `{ "error" : "<error message>" }` <br />
    **Typical reason:** Data format (json body of the Post request) wrong, e.g. missing key for email or password.

  * **Code:** 400 Bad request <br />
    **Content:** `{ "error" : "email address already in use" }` <br />
    **Typical reason:** Email address already used for an other account.

  * **Code:** 500 Internal server error <br />
    **Content:** `{ "error" : "<error message>" }` <br />
    **Typical reason:** Something went wrong during password hashing

* **Sample Call:**

  ```go
    creds := &userCredentials{
      Email:    "your@email.com", // `json:"email"`
      Password: "yourpassword", // `json:"password"`
    }
    payload, err := json.Marshal(creds)
    resp, err := http.Post(user-service-addr + "/v1/user/signup", "application/json", bytes.NewBuffer(payload))
    defer resp.Body.Close()
  ```
* **Notes:**
  * None
