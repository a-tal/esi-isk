import(/* webpackPreload: true */ 'jquery');
import(/* webpackPreload: true */ 'bootstrap');
import(/* webpackPreload: true */ 'github-fork-ribbon-css/gh-fork-ribbon.css');

import Cookie from 'js-cookie';
import { pad } from './utils';

function header() {
  let title = document.createElement('h1');
  title.classList.add('display-1');
  title.classList.add('text-center');

  let titleLink = document.createElement('a');
  titleLink.href = 'javascript:window.m()';
  titleLink.classList.add('text-dark');
  titleLink.innerHTML = 'ESI ISK';
  title.append(titleLink);

  let description = document.createElement('h2');
  description.classList.add('text-muted');
  description.classList.add('text-center');
  description.innerHTML = 'Tracking ISK donations and zero ISK item exchange contracts';

  let header = document.createElement('div');
  header.id = "header";
  header.appendChild(title);
  header.appendChild(description);
  header.classList.add('pb-4');
  return header;
}

// create & return the sign in/up div
function signup(prefs, hasCookie) {
  let details = signupDetails(prefs, hasCookie);
  let signupLink = document.createElement('a');
  signupLink.id = 'signup';
  signupLink.href = details.target;
  signupLink.classList.add('github-fork-ribbon');
  signupLink.classList.add('right-top');
  signupLink.classList.add(details.cls);
  signupLink.innerHTML = details.content;
  signupLink.title = details.content;
  signupLink.setAttribute('data-ribbon', details.content);

  return signupLink;
}

function signupDetails(prefs, hasCookie) {
  let target = "/signup";
  let content = "sign up";
  let cls = "ribbon-red";

  if (hasCookie) {
    if (prefs != undefined) {
      target = 'javascript:window.l()';
      content = "log out";
      cls = "ribbon-black";
    } else {
      target = 'javascript:window.p()';
      content = "preferences";
      cls = "ribbon-green";
    }
  }

  return {
    target: target,
    content: content,
    cls: cls
  };
}

// create & return the top recipients div
function getTop() {
  let top = document.createElement('div');
  top.id = 'top';
  top.classList.add('container');

  let donators = document.createElement('div');
  donators.id = 'donators';
  donators.classList.add('row');

  let recipients = document.createElement('div');
  recipients.id = 'recipients';
  recipients.classList.add('row');

  let recipTitle = document.createElement('h2');
  recipTitle.innerHTML = 'Top Recipients';
  recipTitle.classList.add('col-12');
  recipTitle.classList.add('text-center');
  recipTitle.classList.add('py-4');

  top.appendChild(document.createElement('hr'));
  top.appendChild(recipTitle);
  top.appendChild(recipients);

  let donoTitle = document.createElement('h2');
  donoTitle.innerHTML = 'Top Donators';
  donoTitle.classList.add('col-12');
  donoTitle.classList.add('text-center');
  donoTitle.classList.add('py-4');

  top.appendChild(document.createElement('hr'));
  top.appendChild(donoTitle);
  top.appendChild(donators);

  jQuery.ajax({
    url: "/api/top",
    success: function(t) {
      for (let i = 0; i < t.donators.length; i++) {
        donators.appendChild(characterDiv(
         t.donators[i].id,
         t.donators[i].name,
         t.donators[i].donated_isk || 0
       ))
      }

      for (let i = 0; i < t.recipients.length; i++) {
        recipients.appendChild(characterDiv(
          t.recipients[i].id,
          t.recipients[i].name,
          t.recipients[i].received_isk || 0
        ))
      }
    },
    // XXX handle failure here
  });

  return top;
}

function characterDiv(charID, charName, isk) {
  let div = document.createElement('div');
  div.classList.add('character');
  div.classList.add('col-2');
  div.classList.add('text-center');

  let link = document.createElement('a');
  link.href = "javascript:window.c("+charID+");"

  let char = largeCharacterImage(charID, charName);
  link.appendChild(char);

  let iskLine = document.createElement('p'); // maybe div
  iskLine.classList.add('top-isk');
  iskLine.classList.add('text-muted');
  iskLine.classList.add('text-center');
  iskLine.classList.add('font-weight-light');

  let iskAmount = document.createElement('small');
  iskAmount.innerHTML = formatISK(isk);
  iskLine.appendChild(iskAmount);

  div.appendChild(link);
  div.appendChild(iskLine);

  return div
}

function setSignupBanner() {
  let s = document.getElementById('signup');
  let urlParams = new URLSearchParams(window.location.hash.substr(1));
  let details = signupDetails(urlParams.get('prefs'), loggedIn());

  s.href = details.target;
  s.title = details.content;
  s.innerHTML = details.content;
  s.setAttribute('data-ribbon', details.content);
  s.classList.add(details.cls);

  let availCls = ['ribbon-black', 'ribbon-green', 'ribbon-red'];
  for (let i = 0; i < availCls.length; i++) {
    if (details.cls != availCls[i]) {
      s.classList.remove(availCls[i]);
    }
  }
}

// returns the cleared body div
function clearBodyDiv() {
  setSignupBanner();
  let prevAlert = document.getElementsByClassName('alert');
  for (; prevAlert.length > 0;) {
    prevAlert[0].remove();
  }
  let body = document.getElementById('body');
  for (; body.children.length > 0;) {
    body.children[0].remove();
  }
  return body;
}

// exposed as window.c because reasons
function switchCharacterView(charID) {
  window.location.hash = '#c=' + charID;
  let body = clearBodyDiv();
  body.appendChild(characterViewDiv(charID));
}

// exposed as window.m because reasons
function switchToMainPage() {
  window.location.hash = '';
  let body = clearBodyDiv();
  body.appendChild(getTop());
}

// exposed as window.p because reasons
function switchToPrefs(prefType, charID) {
  prefType = validPrefType(prefType)
  let hash = '#prefs&t=' + prefType;
  if (setCharID(charID)) {
    hash += '&c=' + charID;
  }
  window.location.hash = hash;
  let body = clearBodyDiv();
  body.appendChild(prefsView(prefType, charID));
}

function setCharID(charID) {
  // this is insecure, it's only for display purposes tho
  if (charID != undefined && charID > 0) {
    Cookie.set('charID', charID);
    return true;
  }
  return false;
}

// exposed as window.l because reasons
function logout() {
  delCookie();
  createAlert('You have been logged out', 'success');
  switchToMainPage();
}

function delCookie() {
  Cookie.remove('charID');
  document.cookie = 'esi-isk=; expires=Thu, 01 Jan 1970 00:00:01 GMT;';
}

function frontPage() {
  let body = document.createElement('div');
  body.id = "body";
  body.appendChild(getTop());
  return body;
}

function createTable() {
  let table = document.createElement('table');
  table.classList.add('table');
  table.classList.add('table-hover');
  table.classList.add('table-sm');
  return table;
}

function createTableHeader(content) {
  let th = document.createElement('th');
  th.classList.add('text-center');
  th.innerHTML = content;
  return th;
}


// donations in
function donationsTable() {
  let table = createTable();
  table.id = "donations";

  let header = document.createElement('thead');
  let headerRow = document.createElement('tr');

  headerRow.appendChild(createTableHeader("Donator"));
  headerRow.appendChild(createTableHeader("Amount"));
  headerRow.appendChild(createTableHeader("Note"));
  headerRow.appendChild(createTableHeader("Timestamp"));

  header.appendChild(headerRow);
  table.appendChild(header);

  let body = document.createElement('tbody');
  table.appendChild(body);

  return table;
}

// donations out
function donatedTable() {
  let table = createTable();
  table.id = "donations";

  let header = document.createElement('thead');
  let headerRow = document.createElement('tr');

  headerRow.appendChild(createTableHeader("Receiver"));
  headerRow.appendChild(createTableHeader("Amount"));
  headerRow.appendChild(createTableHeader("Note"));
  headerRow.appendChild(createTableHeader("Timestamp"));

  header.appendChild(headerRow);
  table.appendChild(header);

  let body = document.createElement('tbody');
  table.appendChild(body);

  return table;
}

// contracts in
function contractsTable() {
  let table = createTable()
  table.id = "contracts";

  let header = document.createElement('thead');
  let headerRow = document.createElement('tr');

  headerRow.appendChild(createTableHeader("Donator"));
  headerRow.appendChild(createTableHeader("Value"));
  headerRow.appendChild(createTableHeader("Note"));
  headerRow.appendChild(createTableHeader("Location"));
  headerRow.appendChild(createTableHeader("Accepted"));
  headerRow.appendChild(createTableHeader("Issued"));
  headerRow.appendChild(createTableHeader("Expires"));

  header.appendChild(headerRow);
  table.appendChild(header);

  let body = document.createElement('tbody');
  table.appendChild(body);

  return table;
}

function contractedTable() {
  let table = createTable();
  table.id = "contracts";

  let header = document.createElement('thead');
  let headerRow = document.createElement('tr');

  headerRow.appendChild(createTableHeader("Receiver"));
  headerRow.appendChild(createTableHeader("Value"));
  headerRow.appendChild(createTableHeader("Note"));
  headerRow.appendChild(createTableHeader("Location"));
  headerRow.appendChild(createTableHeader("Accepted"));
  headerRow.appendChild(createTableHeader("Issued"));
  headerRow.appendChild(createTableHeader("Expires"));

  header.appendChild(headerRow);
  table.appendChild(header);

  let body = document.createElement('tbody');
  table.appendChild(body);

  return table;
}

function dayRow(dt, colSpan) {
  let row = document.createElement('tr');
  let day = createTD(dt.toDateString());
  day.classList.add('table-active');
  day.colSpan = colSpan;
  row.appendChild(day);
  return row;
}

function createTD(content='', centered=true) {
  let td = document.createElement('td');
  td.classList.add('align-middle');
  if (centered) {
    td.classList.add('text-center');
  }
  td.innerHTML = content;
  return td;
}

function donationRow(d, donation) {
  let row = document.createElement('tr');

  let contact = createTD();
  let contactLink = document.createElement('a');
  contact.appendChild(contactLink);

  if (donation == true) {
    contactLink.href = "javascript:window.c(" + d.donator + ");"
    contactLink.appendChild(smallCharacterImage(d.donator));
  } else {
    contactLink.href = "javascript:window.c(" + d.receiver + ");"
    contactLink.appendChild(smallCharacterImage(d.receiver));
  }

  let ts = new Date(d.timestamp);

  row.appendChild(contact);
  row.appendChild(createTD(formatISK(d.amount)));
  row.appendChild(createTD(d.note || '', false));
  row.appendChild(createTD(pad(ts.getUTCHours(), 2) + ':' + pad(ts.getUTCMinutes(), 2) + ':' + pad(ts.getUTCSeconds(), 2)));

  return row;
}

function contractRow(d, donation) {
  let contact = createTD();
  let contactLink = document.createElement('a');
  contact.appendChild(contactLink);

  if (donation == true) {
    contactLink.href = "javascript:window.c(" + d.donator + ");"
    contactLink.appendChild(smallCharacterImage(d.donator));
  } else {
    contactLink.href = "javascript:window.c(" + d.receiver + ");"
    contactLink.appendChild(smallCharacterImage(d.receiver));
  }

  let row = document.createElement('tr');

  let ts = new Date(d.issued);

  row.appendChild(contact);
  row.appendChild(createTD(formatISK(d.value)));
  row.appendChild(createTD(d.note));
  row.appendChild(createTD(d.location));
  row.appendChild(createTD(d.accepted));
  row.appendChild(createTD(pad(ts.getUTCHours(), 2) + ':' + pad(ts.getUTCMinutes(), 2) + ':' + pad(ts.getUTCSeconds(), 2)));
  row.appendChild(createTD(d.expires));

  row.classList.add('contract-collapsed');
  row.title = 'click to expand/collapse';
  return row
}

function contractItems(d) {
  let itemsTable = createTable();
  itemsTable.classList.add("d-none");

  let header = document.createElement('thead');
  let headerRow = document.createElement('tr');
  headerRow.classList.add('table-active');

  headerRow.appendChild(createTableHeader("Item"));
  headerRow.appendChild(createTableHeader("Quantity"));

  header.appendChild(headerRow);
  itemsTable.appendChild(header);

  let body = document.createElement('tbody');
  itemsTable.appendChild(body);

  for (let i = 0; i < d.items.length; i++) {
    let itemRow = document.createElement('tr');
    let itemImg = document.createElement('img');

    itemImg.height = 50;
    itemImg.width = 50;
    itemImg.alt = d.items[i].type_id.toString();
    itemImg.classList.add('rounded');
    itemImg.src = 'https://imageserver.eveonline.com/Type/' + d.items[i].type_id + '_64.png';

    let itemTD = createTD();
    itemTD.appendChild(itemImg);
    itemRow.appendChild(itemTD);
    itemRow.appendChild(createTD(d.items[i].quantity));

    body.appendChild(itemRow);
  }

  let itemsTR = document.createElement('tr');
  let itemsTD = document.createElement('td');
  itemsTD.colSpan = 7;
  itemsTD.appendChild(itemsTable);
  itemsTR.appendChild(itemsTD);

  return itemsTR;
}

function formatISK(n) {
  return n.toLocaleString(undefined, {
    maximumFractionDigits: 2,
    minimumFractionDigits: 2
  }) + ' ISK';
}

function largeCharacterImage(charID, charName, imgType, imgExtn, standing) {
  let charImg = document.createElement('img');
  charImg.height = 128;
  charImg.width = 128;
  charImg.alt = charID.toString();
  charImg.classList.add('rounded');
  charImg.src = 'https://imageserver.eveonline.com/' + (imgType || 'Character') + '/' + charID + '_128.' + (imgExtn || 'jpg');

  let charDiv = document.createElement('div');
  let charNameP = document.createElement('p');
  charNameP.classList.add('text-muted');
  charNameP.classList.add('my-0');

  if (standing == true) {
    charNameP.appendChild(standingSpan())
  }

  charNameP.innerHTML = charName;
  charDiv.appendChild(charImg);
  charDiv.appendChild(charNameP);

  return charDiv;
}

function standingSpan() {
  let span = document.createElement('span');
  span.classList.add('text-success');
  span.classList.add('small');
  span.classList.add('pl-1');
  span.innerHTML = '✓';
  span.title = 'user is in good standing';
  return span
}

function smallCharacterImage(charID) {
  let charImg = document.createElement('img');
  charImg.height = 50;
  charImg.width = 50;
  charImg.alt = charID.toString();
  charImg.classList.add('rounded');
  charImg.src = 'https://imageserver.eveonline.com/Character/' + charID + '_64.jpg';
  return charImg;
}

// return title as an h3 with muted and centered text
function h3title(title) {
  let h3 = document.createElement('h3');
  h3.classList.add('text-center');
  h3.classList.add('text-muted');
  h3.innerHTML = title;
  return h3;
}

function characterView(charID) {
  let body = document.createElement('div');
  body.id = "body";
  body.appendChild(characterViewDiv(charID));
  return body
}

function characterViewDiv(charID) {
  let character = document.createElement('div');
  character.classList.add('text-center');
  character.classList.add('pb-4');
  character.classList.add('row');
  character.classList.add('justify-content-center');

  let charImg = largeCharacterImage(charID, '');
  charImg.classList.add('col-2');
  character.appendChild(charImg);

  let details = document.createElement('div');
  details.appendChild(character);

  // in
  let donations = donationsTable();
  let contracts = contractsTable();
  // out
  let donated = donatedTable();
  let contracted = contractedTable();

  let urlParams = new URLSearchParams(window.location.hash.substr(1));
  let passphrase = urlParams.get('p');

  let url = '/api/char?c=' + charID;
  if (passphrase != undefined && passphrase != '') {
    url += '&p=' + passphrase;
  }

  jQuery.ajax({
    url: url,
    success: function(t) {
      charImg.getElementsByTagName('p')[0].innerHTML = t.character.name;
      if (t.character.good_standing == true) {
        charImg.getElementsByTagName('p')[0].appendChild(standingSpan())
      }

      let corpImg = largeCharacterImage(
        t.character.corporation,
        t.character.corporation_name,
        'Corporation',
        'png',
      );
      corpImg.classList.add('col-2');
      character.appendChild(corpImg);

      if (t.character.alliance != undefined) {
        let alianceImg = largeCharacterImage(
          t.character.alliance,
          t.character.alliance_name,
          'Alliance',
          'png',
        );
        alianceImg.classList.add('col-2');
        character.appendChild(alianceImg);
      }

      if (t.contracts != undefined) {
        displayTable(details, contracts, "contracts", t.contracts, contractRow, true, 7)
      }

      if (t.donations != undefined) {
        displayTable(details, donations, "donations", t.donations, donationRow, true)
      }

      if (t.contracted != undefined) {
        displayTable(details, contracted, "contracted", t.contracted, contractRow, false, 7)
      }

      if (t.donated != undefined) {
        displayTable(details, donated, "donated", t.donated, donationRow, false)
      }

    },

    error: function(r, s, e) {
      if (r.status == 403) {
        createAlert('This is a private profile', 'warning');
      } else {
        console.log(r.status + ' ' + e);
        createAlert('Failed to get details for ' + charID, 'warning');
      }
    },
  });

  return details;
}

function displayTable(details, parent, title, array, genFunc, donation, colSpan=4) {
  details.appendChild(h3title(title));
  details.appendChild(parent);
  let day = 0;

  for (let i=0; i < array.length; i++) {
    let thisDay = new Date(array[i].timestamp || array[i].issued);
    if (thisDay.getUTCDate() != day) {
      day = thisDay.getUTCDate();
      parent.tBodies[0].appendChild(dayRow(thisDay, colSpan));
    }
    parent.tBodies[0].appendChild(genFunc(array[i], donation));
    // XXX this is hacky, should be more explicit here
    if (colSpan != 4) {
      parent.tBodies[0].appendChild(contractItems(array[i]));
    }
  }
}

function footer() {
  let copyright = document.createElement('p');
  copyright.innerHTML = 'Source: <a href="https://github.com/a-tal/esi-isk">https://github.com/a-tal/esi-isk</a><br />Released under the MIT license<br />© 2018 Adam Talsma';
  copyright.classList.add('small');
  copyright.classList.add('text-muted');
  copyright.classList.add('text-center');

  let footer = document.createElement('div');
  footer.id = "footer";
  footer.appendChild(document.createElement('hr'));
  footer.appendChild(copyright);
  return footer;
}

function formInput(id, label, name, value, placeholder, details, modifier) {
  let iDiv = document.createElement('div');
  iDiv.classList.add('form-group');

  let iLabel = document.createElement('label');
  iLabel.for = id;
  iLabel.innerHTML = label;
  iDiv.appendChild(iLabel);

  let i = document.createElement('input');
  i.name = name;
  i.value = value || '';
  i.id = id;
  i.classList.add('form-control');
  i.placeholder = placeholder;
  i.setAttribute('aria-describedby', id + '-details');
  if (modifier != undefined) {
   modifier(i);
  }
  iDiv.appendChild(i);

  let iDetails = document.createElement('small');
  iDetails.classList.add('form-text');
  iDetails.classList.add('text-muted');
  iDetails.innerHTML = details;
  iDetails.id = id + '-details';
  iDiv.appendChild(iDetails);

  return iDiv;
}

function formRow(columns, modifier) {
  let row = document.createElement('div');
  row.className = 'row';

  for (let i = 0; i < columns.length; i++) {
   let col = document.createElement('div');
   col.className = 'col';
   if (modifier != undefined) {
     modifier(col)
   }
   col.appendChild(columns[i]);
   row.appendChild(col);
  }

  return row;
}

function validPrefType(prefType) {
  if (prefType != 'c' && prefType != 'd' && prefType != 'a') {
   prefType = 'd';
  }
  return prefType
}

function prefsView(prefType, char) {
  prefType = validPrefType(prefType)

  let charID = char || Cookie.get('charID') || 'charID';

  let container = document.createElement('div');
  container.className = 'container-fluid';
  container.id = 'body';

  let header = document.createElement('h3');
  header.classList.add('text-muted');
  header.classList.add('text-center');

  let url = document.createElement('h4');
  url.classList.add('text-muted');
  url.classList.add('text-center');
  url.classList.add('pt-4');
  url.classList.add('pb-2');
  let link = document.createElement('a');
  let href = document.location.origin + '/api/custom?t=' + prefType + '&c=' + charID;
  link.href = href;
  link.innerHTML = href;
  link.target = "_blank";
  url.appendChild(link);

  let form = document.createElement('form');
  form.id = 'prefs';
  form.action = 'javascript:window.P("' + prefType + '")';

  container.appendChild(header);
  container.appendChild(form);
  container.appendChild(url);

  let typeName = 'Donation';

  let switchType = 'c';
  let switchText = 'Contracts';

  let combinedType = 'a';
  let combinedText = 'Combined';

  switch (prefType) {
    case 'c':
      typeName = 'Contract';
      switchType = 'd';
      switchText = 'Donations';
      break;

    case 'a':
      typeName = 'Combined';
      combinedType = 'c';
      combinedText = 'Contracts'
      switchType = 'd';
      switchText = 'Donations';
      break;
  }

  header.innerHTML = typeName + ' preferences';

  jQuery.ajax({
    url: "/api/prefs?t=" + prefType,
    success: function(t) {
      if (prefType == 'c' || prefType == 'd') {
        buildPrefsForm(form, t);
      } else {
        buildCombinedPrefsForm(form, t);
      }
      addPrefsButtons(form, switchType, switchText, combinedType, combinedText);
      updateExampleURL(link, charID, prefType, t);
    },
    error: function(r, s, e) {
      console.log('Status: ' + s + ' Error: ' + e);
      console.log(r);
      createAlert(
        'Failed to get ' + typeName.toLowerCase() + ' preferences',
        'warning'
      );
    },
  });

  return container;
}

function updateExampleURL(link, charID, prefType, t) {
  let href = document.location.origin + '/api/custom?t=' + prefType + '&c=' + charID;

  let passphrase = t.passphrase;
  if (prefType == 'a') {
    passphrase = t.donations.passphrase;
  }

  if (passphrase != "" && passphrase != undefined) {
    href += '&p=' + passphrase;
  }

  link.href = href;
  link.innerHTML = href;
}

function buildCombinedPrefsForm(form, t) {
  addPrefsHeaderFooter(form, t.donations.header, t.donations.footer);
  form.appendChild(formRow([
    formInput(
      'pref-donation-pattern',
      'Donation pattern',
      'donation_pattern',
      t.donations.pattern,
      'enter a custom row pattern for donations',
      patternHelp(),
    ),
    formInput(
      'pref-contract-pattern',
      'Contract pattern',
      'contract_pattern',
      t.contracts.pattern,
      'enter a custom row pattern for contracts',
      patternHelp(),
    ),
  ]));

  form.appendChild(formRow([
    formInput(
      'pref-rows',
      'Rows',
      'rows',
      t.donations.rows,
      '',
      'number of rows to display',
      function (i) {
        i.type = 'number';
        i.min = 1;
        i.step = 1;
      },
    ),
    formInput(
      'pref-donation-minimum',
      'Donation minimum',
      'donation_minimum',
      t.donations.minimum,
      '',
      'Minimum ISK value to include for donations',
      function (i) {
        i.type = 'number';
        i.min = 0;
        i.step = 0.1;
      },
    ),
    formInput(
      'pref-contract-minimum',
      'Contract minimum',
      'contract_minimum',
      t.contracts.minimum,
      '',
      'Minimum ISK value to include for contracts',
      function (i) {
        i.type = 'number';
        i.min = 0;
        i.step = 0.1;
      },
    )
  ]));

  addPrefsMaxAgePassphrase(form, t.donations.max_age, t.donations.passphrase);
}

function patternHelp() {
  return 'HTML will be stripped. <a href="https://github.com/a-tal/esi-isk/blob/master/README.md#Formatting">Formatting help</a>';
}

function buildPrefsForm(form, t) {
  addPrefsHeaderFooter(form, t.header, t.footer);

  form.appendChild(formRow([formInput(
    'pref-pattern',
    'Pattern',
    'pattern',
    t.pattern,
    'enter a custom row pattern',
    patternHelp(),
  )]));

  form.appendChild(formRow([
    formInput(
      'pref-rows',
      'Rows',
      'rows',
      t.rows,
      '',
      'number of rows to display',
      function (i) {
        i.type = 'number';
        i.min = 1;
        i.step = 1;
      },
    ),
    formInput(
      'pref-minimum',
      'Minimum',
      'minimum',
      t.minimum,
      '',
      'Minimum ISK value to include',
      function (i) {
        i.type = 'number';
        i.min = 0;
        i.step = 0.1;
      },
    )
  ]));

  addPrefsMaxAgePassphrase(form, t.max_age, t.passphrase);
}

function addPrefsMaxAgePassphrase(form, maxAge, passphrase) {
  form.appendChild(formRow([
    formInput(
      'pref-max-age',
      'Max Age',
      'max_age',
      maxAge,
      '',
      'Maximum age of donations to include (seconds)',
      function (i) {
        i.type = 'number';
        i.min = 1;
        i.step = 1;
      },
    ),
    formInput(
      'pref-passphrase',
      'Passphrase',
      'passphrase',
      passphrase,
      'optional passphrase to require',
      'Optional passphrase',
    )
  ]));
}

function addPrefsButtons(form, switchType, switchText, combinedType, combinedText) {
  let link = document.createElement('a');
  link.classList.add('btn');
  link.classList.add('btn-primary');
  link.href = 'javascript:window.p("' + switchType + '");';
  link.setAttribute('role', 'button');
  link.innerHTML = switchText;

  let combined = document.createElement('a');
  combined.classList.add('btn');
  combined.classList.add('btn-primary');
  combined.href = 'javascript:window.p("' + combinedType + '");';
  combined.setAttribute('role', 'button');
  combined.innerHTML = combinedText;

  let button = document.createElement('button');
  button.type = 'submit';
  button.innerHTML = 'Submit';
  button.classList.add("btn");
  button.classList.add("btn-primary");

  let count = 0;
  form.appendChild(formRow([link, combined, button], function (c) {
    c.classList.add('d-flex');
    if (count > 1) {
      c.classList.add('justify-content-end');
    } else if (count > 0) {
      c.classList.add('justify-content-center');
    } else {
      c.classList.add('justify-content-start');
    }
    count++;
  }));
}

function addPrefsHeaderFooter(form, header, footer) {
  form.appendChild(formRow([
    formInput(
      'pref-header',
      'Header',
      'header',
      header,
      'enter a custom header',
      'HTML will be stripped',
    ),
    formInput(
      'pref-footer',
      'Footer',
      'footer',
      footer,
      'enter a custom footer',
      'HTML will be stripped',
    )
  ]));
}

function buildPrefsPost(prefType) {
  if (prefType == 'a') {
    return combinedPrefsPost()
  } else {
    return singularPrefsPost()
  }
}

function combinedPrefsPost() {
  let data = {
    "donations": {
      "pattern": document.getElementById('pref-donation-pattern').value,
      "rows": parseInt(document.getElementById('pref-rows').value),
      "minimum": parseFloat(document.getElementById('pref-donation-minimum').value),
      "max_age": parseInt(document.getElementById('pref-max-age').value)
    },
    "contracts": {
      "pattern": document.getElementById('pref-contract-pattern').value,
      "minimum": parseFloat(document.getElementById('pref-contract-minimum').value)
    }
  }

  let header = document.getElementById('pref-header').value;
  let footer = document.getElementById('pref-footer').value;
  let passphrase = document.getElementById('pref-passphrase').value

  if (header != undefined && header != '') {
    data.donations["header"] = header;
  }
  if (footer != undefined && footer != '') {
    data.donations["footer"] = footer;
  }
  if (passphrase != undefined && passphrase != '') {
    data.donations["passphrase"] = passphrase;
  }

  return data
}

function singularPrefsPost() {
  let data = {
    "pattern": document.getElementById('pref-pattern').value,
    "rows": parseInt(document.getElementById('pref-rows').value),
    "minimum": parseFloat(document.getElementById('pref-minimum').value),
    "max_age": parseInt(document.getElementById('pref-max-age').value)
  };

  let header = document.getElementById('pref-header').value;
  let footer = document.getElementById('pref-footer').value;
  let passphrase = document.getElementById('pref-passphrase').value

  if (header != undefined && header != '') {
    data["header"] = header;
  }
  if (footer != undefined && footer != '') {
    data["footer"] = footer;
  }
  if (passphrase != undefined && passphrase != '') {
    data["passphrase"] = passphrase;
  }

  return data
}

function postPrefs(prefType) {
  prefType = validPrefType(prefType)

  let typeName = 'Donation';
  if (prefType == 'c') {
    typeName = 'Contract';
  } else if (prefType == 'a') {
    typeName = 'Combined';
  }

  $.ajax({
    type: "POST",
    url: "/api/prefs?t=" + prefType,
    data: JSON.stringify(buildPrefsPost(prefType)),
    success: function() {
      createAlert(typeName + ' preferences saved');
      switchToPrefs(prefType);
    },
    error: function(r, s, e) {
      console.log('Status: ' + s + ' Error: ' + e);
      console.log(r);
      createAlert(
        'Failed to save ' + typeName.toLowerCase() + ' preferences',
        'warning'
      );
    },
    dataType: "json",
    contentType: "application/json"
  });
}

function closeSpan() {
 let span = document.createElement('span');
 span.setAttribute('aria-hidden', 'true');
 span.innerHTML = '&times;';
 return span;
}

function createAlert(msg, cls) {
  let prevAlert = document.getElementsByClassName('alert');
  let alertCls = cls || 'primary';

  if (prevAlert.length < 1) {
    let alert = document.createElement('div');
    alert.classList.add('alert');
    alert.classList.add('alert-' + alertCls);
    alert.classList.add('alert-dismissible');
    alert.classList.add('fade');
    alert.classList.add('show');
    alert.classList.add('position-fixed');
    alert.classList.add('w-25');
    alert.classList.add('mt-1');
    alert.setAttribute('role', 'alert');
    alert.innerHTML = msg;

    let closeBtn = document.createElement('button');
    closeBtn.type = 'button';
    closeBtn.className = 'close';
    closeBtn.setAttribute('data-dismiss', 'alert');
    closeBtn.setAttribute('aria-label', 'Close');

    closeBtn.appendChild(closeSpan());
    alert.appendChild(closeBtn);

    let container = document.getElementById('container');
    container.insertBefore(alert, container.childNodes[0]);
  } else {
    let alert = prevAlert[0];
    alert.innerText = msg;
    if (!alert.classList.contains('alert-' + alertCls)) {
      alert.classList.add('alert-' + alertCls)
      let avail = ['primary', 'secondary', 'success', 'danger', 'warning',
                   'info', 'light', 'dark'];
      for (let i = 0; i < avail.length; i++) {
        if (avail[i] != alertCls) {
          alert.classList.remove('alert-' + avail[i]);
        }
      }
    }
  }
}

// NB: this does not confirm the user isn't a hax0rman
function loggedIn() {
  if (document.cookie != undefined) {
    return document.cookie.split('=')[0] == 'esi-isk';
  }
  return false;
}

function content() {
  let container = document.createElement('div');
  container.id = "container";
  container.className = 'container-fluid';

  let urlParams = new URLSearchParams(window.location.hash.substr(1));
  let prefs = urlParams.get('prefs');
  let logout = urlParams.get('logout');
  let hasCookie = loggedIn();

  container.appendChild(signup(prefs, hasCookie));
  container.appendChild(header());

  if (logout != undefined && hasCookie) {
    delCookie();
    hasCookie = false;
  }

  if (prefs != undefined && hasCookie) {
    let charID = urlParams.get('c');
    setCharID(charID);
    container.appendChild(prefsView(urlParams.get('t'), charID));
  } else {
    var charID = urlParams.get('c');
    if (charID != undefined) {
      container.appendChild(characterView(charID));
    } else {
      container.appendChild(frontPage());
    }
  }

  container.appendChild(footer());
  return container;
}

jQuery(function($){
  document.body.removeChild(document.getElementById("nojs"));
  document.body.appendChild(content());
  window.c = switchCharacterView;
  window.m = switchToMainPage;
  window.p = switchToPrefs;
  window.P = postPrefs;
  window.l = logout;

  $(document).on("click", ".contract-collapsed", function(e) {
    let target = e.target;
    while (target.tagName != "TR") {
      target = target.parentElement;
    }
    let table = target.nextSibling.firstElementChild.firstElementChild;
    table.classList.remove('d-none');
    table.classList.add('d-table');
    target.classList.remove("contract-collapsed");
    target.classList.add("contract-expanded");
  });

  $(document).on("click", ".contract-expanded", function(e) {
    let target = e.target;
    while (target.tagName != "TR") {
      target = target.parentElement;
    }
    let table = target.nextSibling.firstElementChild.firstElementChild;
    table.classList.add('d-none');
    table.classList.remove('d-table');
    target.classList.add("contract-collapsed");
    target.classList.remove("contract-expanded");
  });
});
