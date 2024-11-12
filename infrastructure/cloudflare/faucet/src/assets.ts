const styleCSS = `
  body {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100vh;
    margin: 0;
    font-family: Arial, sans-serif;
    background-color: #f0f0f0;
  }
  
  .container {
      text-align: center;
  }
  
  .logo {
      width: 250px;
      margin-bottom: 50px;
  }
  
  form {
      display: flex;
      flex-direction: column;
      justify-content: center;
      gap: 10px;
  }
  
  input[type="text"] {
      padding: 8px;
      font-size: 16px;
  }
  
  button {
      padding: 8px 16px;
      font-size: 16px;
      cursor: pointer;
      background-color: #fc004c;
      color: white;
      border: none;
      border-radius: 4px;
  }
`

const logoSVGBase64 = "PHN2ZyB3aWR0aD0nMTc0JyBoZWlnaHQ9JzI0JyB2aWV3Qm94PScwIDAgMTc0IDI0JyBmaWxsPSdub25lJyB4bWxucz0naHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmcnPjxnIGNsaXAtcGF0aD0ndXJsKCNjbGlwMF82MzAxXzc0NzApJz48cGF0aCBkPSdNOTkuODMgMTMuNDYyNUM5OC42MzU2IDExLjE3NjUgOTcuOTU1OCA5LjU0MjkzIDk3LjA1MiA3LjI1Njk2SDk2LjkyNDFDOTcuMTE4NyA5LjcwNzM3IDk3LjIxNDcgMTEuMzQxIDk3LjIxNDcgMTMuNzg4N1YyMy40MjMySDkzLjAxNTdWMC41NjA3OTFIOTcuNjk5OUwxMDIuMDkzIDkuMjQ5MUMxMDMuMjg4IDExLjU5OTggMTA0IDEzLjEzNjMgMTA0Ljk2NyAxNS40MjIzQzEwNS45MzggMTMuMTM2MyAxMDYuNjQ3IDExLjU5OTggMTA3Ljg0MSA5LjI0OTFMMTEyLjIzNSAwLjU2MDc5MUgxMTYuOTE5VjIzLjQyNTlIMTEyLjcyVjEzLjc5MTRDMTEyLjcyIDExLjM0MSAxMTIuODE2IDkuNzA3MzcgMTEzLjAxMSA3LjI1OTY2SDExMi44ODNDMTExLjk3OSA5LjU0NTYzIDExMS4yOTkgMTEuMTc5MiAxMTAuMTA1IDEzLjQ2NTJMMTA0Ljk2NyAyMy40Mjg2TDk5LjgzIDEzLjQ2NTJWMTMuNDYyNVonIGZpbGw9JyNGRjAwNEQnLz48cGF0aCBkPSdNMTM3LjM2NCAxOC41MjVDMTM2LjQyOCAyMS40NjMzIDEzMy42ODIgMjMuOTEzNyAxMjkuMzIxIDIzLjkxMzdDMTIzLjk2IDIzLjkxMzcgMTIwLjU5OCAxOS45OTQyIDEyMC41OTggMTQuOTMxNkMxMjAuNTk4IDkuODY5MDQgMTI0LjA1MyA1Ljk0OTQ2IDEyOS4yMjIgNS45NDk0NkMxMzMuNDg1IDUuOTQ5NDYgMTM2LjA3MSA4LjIzNTQzIDEzNy4xMzggMTEuNjMyQzEzNy41MjQgMTIuODcyMSAxMzcuNjg3IDE0LjMxMTYgMTM3LjY4NyAxNS43NDg0VjE2LjQwMDhIMTI0Ljc2NUMxMjQuODkzIDE4LjUyNSAxMjYuNDEyIDIwLjgxMSAxMjkuMzE4IDIwLjgxMUMxMzEuNTE1IDIwLjgxMSAxMzIuNjQ1IDE5LjU3MDkgMTMzLjAzMiAxOC41MjVIMTM3LjM2MkgxMzcuMzY0Wk0xMjQuNzY3IDEzLjI5OEgxMzMuNTIzQzEzMy40MjcgMTEuMDEyIDEzMS45MDcgOS4wNTIyNCAxMjkuMjI1IDkuMDUyMjRDMTI2LjU0MyA5LjA1MjI0IDEyNS4wMjYgMTEuMDEyIDEyNC43NjcgMTMuMjk4WicgZmlsbD0nI0ZGMDA0RCcvPjxwYXRoIGQ9J00xNTQuNDI0IDIwLjE1ODdWMjMuNDI1OUgxMzkuNTY0VjIwLjE1ODdMMTQ4Ljk2NCA5LjcwNzRIMTM5Ljg4NlY2LjQ0MDE5SDE1NC4wOThWOS43MDc0TDE0NC42OTggMjAuMTU4N0gxNTQuNDIxSDE1NC40MjRaJyBmaWxsPScjRkYwMDREJy8+PHBhdGggZD0nTTE2NS4yMTMgMjMuOTEzN0MxNTkuODg0IDIzLjkxMzcgMTU2LjQyNiAxOS45OTQyIDE1Ni40MjYgMTQuOTMxNkMxNTYuNDI2IDkuODY5MDQgMTU5Ljg4NCA1Ljk0OTQ2IDE2NS4yMTMgNS45NDk0NkMxNzAuNTQyIDUuOTQ5NDYgMTc0IDkuODY5MDQgMTc0IDE0LjkzMTZDMTc0IDE5Ljk5NDIgMTcwLjU0NSAyMy45MTM3IDE2NS4yMTMgMjMuOTEzN1pNMTY1LjIxMyAyMC42NDkyQzE2OC42MDQgMjAuNjQ5MiAxNjkuOTkzIDE3LjcxMDkgMTY5Ljk5MyAxNC45MzQzQzE2OS45OTMgMTIuMTU3NyAxNjguNjA0IDkuMjE5MzcgMTY1LjIxMyA5LjIxOTM3QzE2MS44MjIgOS4yMTkzNyAxNjAuNDMzIDEyLjE2MDQgMTYwLjQzMyAxNC45MzQzQzE2MC40MzMgMTcuNzA4MiAxNjEuODIyIDIwLjY0OTIgMTY1LjIxMyAyMC42NDkyWicgZmlsbD0nI0ZGMDA0RCcvPjxwYXRoIGQ9J00yMC4yNDU0IDIwLjUwOUwyOC42NDMyIDEyLjAxNzVWMTEuOTc5OEwzNy4wMDEgMjAuNDcxM0MzOS40MTY0IDIyLjkxMzYgNDIuMzcwMyAyNCA0NS4yODQyIDI0QzUxLjM4MTMgMjQgNTcuMjQ2NCAxOS4xOTM1IDU3LjI0NjQgMTEuOTc5OEw2NS42MDQyIDIwLjQ3MTNDNjguMDE5NiAyMi45MTM2IDcwLjk3MzUgMjQgNzMuODg3NCAyNEM3OS45ODQ1IDI0IDg1Ljg0OTYgMTkuMTkzNSA4NS44NDk2IDExLjk3OThINzguNjc4Mkw3MC4zMjAzIDMuNTI4N0M2Ny45MDUgMS4wODYzOCA2NC45MTM4IDAgNjEuOTk5OCAwQzU1LjkwMjggMCA1MC4wNzUgNC43Njg3MyA1MC4wNzUgMTEuOTc5OEw0MS43MTcxIDMuNTI4N0MzOS4zMDE4IDEuMDg2MzggMzYuMzEwNSAwIDMzLjM5NjYgMEMyNy4yOTk2IDAgMjEuNDcxNyA0Ljc2ODczIDIxLjQ3MTcgMTEuOTc5OEwwIDEyLjAxNzVDMCAxOS4yNjkgNS45NDI0NSAyMy45NTk2IDEyLjAzOTUgMjMuOTU5NkMxNC45NTM0IDIzLjk1OTYgMTcuOTA3MyAyMi44NzMyIDIwLjI0NTQgMjAuNTA5WicgZmlsbD0nI0ZGMDA0RCcvPjwvZz48ZGVmcz48Y2xpcFBhdGggaWQ9J2NsaXAwXzYzMDFfNzQ3MCc+PHJlY3Qgd2lkdGg9JzE3NCcgaGVpZ2h0PScyNCcgZmlsbD0nd2hpdGUnLz48L2NsaXBQYXRoPjwvZGVmcz48L3N2Zz4="

export const indexHTML = (turnstileSiteKey: string) => `
  <!DOCTYPE html>
  <html lang="en">
  <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>Mezo Faucet</title>
      <style>
          ${styleCSS}
      </style>
      <script src="https://challenges.cloudflare.com/turnstile/v0/api.js" async defer></script>
  </head>
  <body>
  <div class="container">
      <img src="data:image/svg+xml;base64,${logoSVGBase64}" alt="Logo" class="logo">
  
      <form action="/" method="POST">
          <input type="text" name="address" placeholder="0x..." required>
          <button type="submit">Request BTC</button>
          <div class="cf-turnstile" data-sitekey="${turnstileSiteKey}"></div>
      </form>
  </div>
  </body>
  </html>
`

export const errorHTML = (error: string) => `
  <!DOCTYPE html>
  <html lang="en">
  <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>Mezo Faucet</title>
      <style>
          ${styleCSS}
      </style>
  </head>
  <body>
  <div class="container">
      <img src="data:image/svg+xml;base64,${logoSVGBase64}" alt="Logo" class="logo">
      
      <h1 style="color: red">Your request has been rejected!</h1>  
      <p><b>Cause: </b>${error}</p>
  </div>
  </body>
  </html>
`

export const successHTML = (txHash: string, amountBTC: string) => `
  <!DOCTYPE html>
  <html lang="en">
  <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>Mezo Faucet</title>
      <style>
          ${styleCSS}
      </style>
  </head>
  <body>
  <div class="container">
      <img src="data:image/svg+xml;base64,${logoSVGBase64}" alt="Logo" class="logo">
      
      <h1 style="color: green">${amountBTC} BTC sent!</h1>  
      <p><b>Transaction: </b><a href="https://explorer.test.mezo.org/tx/${txHash}">${txHash}</a></p>
  </div>
  </body>
  </html>
`