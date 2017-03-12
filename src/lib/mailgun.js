import request from 'request';

/**
 * Send an email to a single recipient.
 * @param {string} domain
 * @param {string} from
 * @param {string} to
 * @param {string} subject
 * @param {string} text
 */
export function send_email(domain, from, to, subject, text) {

  const mailgun_api_key = process.env.MAILGUN_API_KEY;
  if (!mailgun_api_key) {
    console.error('warning: no mailgun api key');
  }

  const form = {
    from,
    to,
    subject,
    text
  };

  const params = {
    uri: `https://api.mailgun.net/v3/${domain}/messages`,
    method: 'POST',
    auth: {
      user: 'api',
      password: mailgun_api_key
    },
    form
  };

  return new Promise((resolve, reject) => {
    request(params, (err, res, body) => {
      if (err) reject(err);
      else resolve(body);
    });
  });
}
