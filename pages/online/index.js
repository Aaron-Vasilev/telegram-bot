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

async function boot() {
  if (!telegramUserId) {
    setStatus('Open this page through Telegram to subscribe.', 'error')
    return
  }

  let config
  try {
    const res = await fetch('/api/online-config')
    if (!res.ok) throw new Error(`config ${res.status}`)
    config = await res.json()
  } catch (e) {
    setStatus('Could not load payment options. Please try again later.', 'error')
    return
  }

  if (!config.clientId || !config.planId) {
    setStatus('Payment not configured. Contact admin.', 'error')
    return
  }

  const sdk = document.createElement('script')
  sdk.src = `https://www.paypal.com/sdk/js?client-id=${encodeURIComponent(config.clientId)}&vault=true&intent=subscription`
  sdk.onload = () => renderButton(config.planId)
  sdk.onerror = () => setStatus('PayPal failed to load.', 'error')
  document.head.appendChild(sdk)
}

function renderButton(planId) {
  window.paypal.Buttons({
    style: { shape: 'pill', color: 'blue', layout: 'vertical', label: 'subscribe' },
    createSubscription: (_data, actions) => actions.subscription.create({
      plan_id: planId,
      custom_id: telegramUserId,
    }),
    onApprove: async (data) => {
      setStatus('Activating your subscription…')
      try {
        const res = await fetch('/api/subscription-success', {
          method: 'POST',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body: new URLSearchParams({
            subscription_id: data.subscriptionID,
            telegram_user_id: telegramUserId,
          }),
        })
        if (!res.ok) throw new Error(`activate ${res.status}`)
        setStatus('✅ Subscription activated! See you on the mat 🧘', 'success')
      } catch (e) {
        setStatus('Payment received but activation failed. Contact @vialettochka.', 'error')
      }
    },
    onError: () => setStatus('Something went wrong. Please try again.', 'error'),
  }).render('#paypal-button-container')
}

boot()
