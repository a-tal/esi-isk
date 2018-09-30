# ESI ISK

ISK donations and zero ISK contract tracking. Live at https://isk.a-t.al/

Once logged in with EVE SSO, ESI ISK will monitor your character's wallet and contracts for donations and zero ISK contracts. You can then use ESI ISK to create a custom HTML output of this activity and embed it in your streams or elsewhere.

The service is free to use, if you feel like donating you can to the character `Send ISK Thanks`.


# Custom API Docs

The custom API response is built using your preferences. In general, you can provide a header, a template for each row of the response (different for contracts vs donations) and a footer. Your content will be html escaped, you are advised to use local css for styling.

The URL is `/api/custom`, the following query string arguments are accepted:

Argument | Meaning      | Default
---------|--------------|-------
`c`      | Character ID |
`t`      | Type, one of `d` for donations, `c` for contracts, or `a` for all | `d`
`p`      | Passphrase, if locked and in good standing |

An auto-refresh is included for your overlay embedding needs.

XXX: if anyone comes up with a decent default style they would like included let me know.


## Passphrases

As noted, your account must be in good standing in order to use passphrases. By using a passphrase you can enable a rudimentary amount of security in keeping your donation history private, if you so choose.

In order to maintain an account in good standing, 1+% of ISK received (donations and value of accepted zero ISK contracts) should be donated to `Send ISK Thanks`. Contracted items do not count towards standing.

Note that setting a passphrase on your donation preferences will also set that same passphrase on your character details (`/api/chars`). Each view (donation, contracts, combined) can have its own passphrase.


## Formatting

The following row template keywords are available for you to use:

Keyword        | Content | Example
---------------|---------|--------
%NAME%         | Your character's name |
%CHARACTER%    | The name of the gifting character |
%AMOUNT%       | The amount/value of ISK donated (with cents) | 10,000,000.00
%AMOUNTISK%    | The amount/value of ISK donated | 10,000,000
%AMOUNTRAW%    | The amount/value of ISK donated (with cents, no commas) | 10000000.00
%AMOUNTRAWISK% | The amount/value of ISK donated (no commas) | 10000000
%DAY%          | The day of the donation | 25
%DAYSUFFIX%    | The two letter suffix for the date | th
%MONTH%        | The date of the donation | Dec
%MONTHLONG%    | The date of the donation with the month spelled out | December
%YEAR%         | The year of the donation | 2018
%TIME%         | The time of the donation | 22:34
%TIMEFULL%     | The time of the donation with the second | 22:34:56
%TIMEAMPM%     | The time of the donation with am/pm | 10:34 PM
%TIMEFULLAMPM% | The time of the donation with the second and am/pm | 10:34:56 PM
%ISODATE%      | ISO3339 standard datetime | 2018-12-25T22:34:50Z
%NOTE%         | Message provided with the donation | Hello, world
%ITEMS%        | Number of items contracted (contracts only) | 42
