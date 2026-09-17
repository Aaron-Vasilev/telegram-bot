const isDev = window.location.host !== 'yogabot.onedoq.com'

const tg = window.Telegram?.WebApp
if (tg) tg.ready()

const user = tg?.initDataUnsafe?.user

if (user?.first_name) {
  const fullName = [user.first_name, user.last_name].filter(Boolean).join(' ')
  document.getElementById('user-greeting').textContent = `Hello, ${fullName} 👋`
}

const urlParams = new URLSearchParams(window.location.search)
const telegramUserIdFromUrl = urlParams.get('telegram_user_id')

let telegramUserId = ''
if (user?.id) {
  telegramUserId = String(user.id)
} else if (telegramUserIdFromUrl) {
  telegramUserId = telegramUserIdFromUrl
} else if (isDev) {
  telegramUserId = '362575139'
}

document.getElementById('telegram-user-id').value = telegramUserId

const status = document.getElementById('status-message')

function setStatus(text, type) {
  status.textContent = text
  status.className = 'status-message' + (type ? ' ' + type : '')
}

function renderButton() {
  const container = document.getElementById('subscribe-container')
  container.innerHTML = ''

  const btn = document.createElement('button')
  btn.type = 'submit'
  btn.textContent = 'Subscribe — 111₪/month'
  btn.className = 'subscribe-btn'
  container.appendChild(btn)
}

const customerForm = document.getElementById('customer-form')
customerForm.addEventListener('submit', (e) => {
  e.preventDefault()
  startCheckout()
})

async function startCheckout() {
  if (!telegramUserId) {
    setStatus('Open this page through Telegram to subscribe.', 'error')
    return
  }

  const customerName = document.getElementById('customer-name').value.trim()
  const customerPhone = document.getElementById('customer-phone').value.trim()
  const termsAgreed = document.getElementById('terms-agree').checked

  if (!customerName || !customerPhone) {
    setStatus('Please fill in your name and phone number.', 'error')
    return
  }
  if (!termsAgreed) {
    setStatus('Please agree to the Terms & Conditions.', 'error')
    return
  }

  setStatus('Redirecting to payment…')
  try {
    const res = await fetch('/api/create-subscription', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        telegram_user_id: telegramUserId,
        customer_name: customerName,
        customer_phone: customerPhone,
      }),
    })
    if (!res.ok) throw new Error(`create ${res.status}`)
    const data = await res.json()
    if (data.redirect_url) {
      window.location.href = data.redirect_url
    } else {
      throw new Error('no redirect_url')
    }
  } catch (e) {
    setStatus('Could not start checkout. Please try again.', 'error')
  }
}

renderButton()

if (urlParams.get('failed')) {
  setStatus('Payment was not completed. Please try again.', 'error')
}
