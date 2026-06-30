const prices = {
  membership_1: '280₪',
  membership_2: '400₪',
  single_first: '70₪',
  single: '90₪',
}

const tg = window.Telegram?.WebApp
if (tg) tg.ready()

const user = tg?.initDataUnsafe?.user

// Show user's full name in greeting
if (user?.first_name) {
  const fullName = [user.first_name, user.last_name].filter(Boolean).join(' ')
  document.getElementById('user-greeting').textContent = `Hello, ${fullName} 👋`
}

if (user?.id) {
  document.getElementById('telegram-user-id').value = user.id
} else if (process.env.NODE_ENV === 'development') {
  document.getElementById('telegram-user-id').value = '362575139'
}

const btn = document.getElementById('submit-btn')

document.querySelectorAll('.plan-card input[type="radio"]').forEach(radio => {
  radio.addEventListener('change', () => {
    btn.textContent = prices[radio.value]
      ? `Proceed to payment — ${prices[radio.value]}`
      : 'Proceed to payment'
  })
})

document.getElementById('payment-form').addEventListener('submit', async e => {
  e.preventDefault()

  const planEl = document.querySelector('[name="plan"]:checked')
  let telegramUserId = document.getElementById('telegram-user-id').value

  if (process.env.NODE_ENV === 'development') {
    telegramUserId = '362575139'
  }

  if (!planEl) return

  if (!telegramUserId) {
    alert('Could not identify your Telegram account. Please open this page through Telegram.')
    return
  }

  btn.disabled = true
  btn.textContent = 'Processing…'

  try {
    const apiBase = process.env.PAYMENT_API_URL || ''
    const res = await fetch(`${apiBase}/api/create-payment`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({ plan: planEl.value, telegram_user_id: telegramUserId }),
    })

    if (!res.ok) throw new Error(`Server error ${res.status}`)

    const data = await res.json()
    if (data.redirect_url) window.location.href = data.redirect_url
  } catch {
    btn.disabled = false
    btn.textContent = prices[planEl.value]
      ? `Proceed to payment — ${prices[planEl.value]}`
      : 'Proceed to payment'
    alert('Something went wrong. Please try again.')
  }
})
