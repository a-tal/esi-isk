import(/* webpackPreload: true */ 'jquery');
import(/* webpackPreload: true */ 'bootstrap');
import(/* webpackPreload: true */ 'bootstrap/dist/css/bootstrap.min.css');
import(/* webpackPreload: true */ 'github-fork-ribbon-css/gh-fork-ribbon.css');


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
function signup() {
  let signupLink = document.createElement('a');
  signupLink.id = 'signup';
  signupLink.href = '/signup';
  signupLink.classList.add('github-fork-ribbon');
  signupLink.classList.add('right-top');
  signupLink.innerHTML = 'sign up';
  signupLink.title = 'sign up';
  signupLink.setAttribute('data-ribbon', 'sign up');

  return signupLink;
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
      for (i = 0; i < t.donators.length; i++) {
        donators.appendChild(characterDiv(
         t.donators[i].id,
         t.donators[i].name,
         t.donators[i].donated_isk || 0
       ))
      }

      for (i = 0; i < t.recipients.length; i++) {
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

// returns the cleared body div
function clearBodyDiv() {
  let body = document.getElementById('body');
  for (i = 0; i <= body.children.length; i++) {
    body.children[0].remove();
  }
  return body;
}

// exposed as window.c because reasons
function switchCharacterView(charID) {
  let body = clearBodyDiv();
  body.appendChild(characterViewDiv(charID));
  window.history.pushState(window.history.state, "ESI ISK - " + charID, '/?c=' + charID);
}

// exposed as window.m because reasons
function switchToMainPage() {
  let body = clearBodyDiv();
  body.appendChild(getTop());
  window.history.pushState(window.history.state, "ESI ISK", '/');
}

function frontPage() {
  let body = document.createElement('div');
  body.id = "body";
  body.appendChild(getTop());
  return body;
}

// donations in
function donationsTable() {
  let table = document.createElement('table');
  table.id = "donations";
  table.classList.add("table");

  let header = document.createElement('thead');
  let headerRow = document.createElement('tr');

  let donator = document.createElement('th');
  donator.innerHTML = "Donator";

  let amount = document.createElement('th');
  amount.innerHTML = "Amount";

  let note = document.createElement('th');
  note.innerHTML = "Note";

  let timestamp = document.createElement('th');
  timestamp.innerHTML = "Timestamp";

  headerRow.appendChild(donator);
  headerRow.appendChild(amount);
  headerRow.appendChild(note);
  headerRow.appendChild(timestamp);

  header.appendChild(headerRow);
  table.appendChild(header);

  let body = document.createElement('tbody');
  table.appendChild(body);

  return table;
}

// donations out
function donatedTable() {
  let table = document.createElement('table');
  table.id = "donations";
  table.classList.add("table");

  let header = document.createElement('thead');
  let headerRow = document.createElement('tr');

  let receiver = document.createElement('th');
  receiver.innerHTML = "Receiver";

  let amount = document.createElement('th');
  amount.innerHTML = "Amount";

  let note = document.createElement('th');
  note.innerHTML = "Note";

  let timestamp = document.createElement('th');
  timestamp.innerHTML = "Timestamp";

  headerRow.appendChild(receiver);
  headerRow.appendChild(amount);
  headerRow.appendChild(note);
  headerRow.appendChild(timestamp);

  header.appendChild(headerRow);
  table.appendChild(header);

  let body = document.createElement('tbody');
  table.appendChild(body);

  return table;
}

// contracts in
function contractsTable() {
  let table = document.createElement('table');
  table.id = "contracts";
  table.classList.add("table");

  let header = document.createElement('thead');
  let headerRow = document.createElement('tr');

  let donator = document.createElement('th');
  donator.innerHTML = "Donator";

  let location = document.createElement('th');
  location.innerHTML = "Location";

  let system = document.createElement('th');
  system.innerHTML = "System";

  let accepted = document.createElement('th');
  accepted.innerHTML = "Accepted";

  let issued = document.createElement('th');
  issued.innerHTML = "Issued";

  let expired = document.createElement('th');
  expired.innerHTML = "Expired";

  // XXX ADD CONTRACT ITEMS -- NESTED ON CLICK IDEALLY

  headerRow.appendChild(donator);
  headerRow.appendChild(location);
  headerRow.appendChild(system);
  headerRow.appendChild(accepted);
  headerRow.appendChild(issued);
  headerRow.appendChild(expired);

  header.appendChild(headerRow);
  table.appendChild(header);

  let body = document.createElement('tbody');
  table.appendChild(body);

  return table;
}

function contractedTable() {
  let table = document.createElement('table');
  table.id = "contracts";
  table.classList.add("table");

  let header = document.createElement('thead');
  let headerRow = document.createElement('tr');

  let receiver = document.createElement('th');
  receiver.innerHTML = "Receiver";

  let location = document.createElement('th');
  location.innerHTML = "Location";

  let system = document.createElement('th');
  system.innerHTML = "System";

  let accepted = document.createElement('th');
  accepted.innerHTML = "Accepted";

  let issued = document.createElement('th');
  issued.innerHTML = "Issued";

  let expired = document.createElement('th');
  expired.innerHTML = "Expired";

  // XXX ADD CONTRACT ITEMS -- NESTED ON CLICK IDEALLY

  headerRow.appendChild(receiver);
  headerRow.appendChild(location);
  headerRow.appendChild(system);
  headerRow.appendChild(accepted);
  headerRow.appendChild(issued);
  headerRow.appendChild(expired);

  header.appendChild(headerRow);
  table.appendChild(header);

  let body = document.createElement('tbody');
  table.appendChild(body);

  return table;
}

function donationRow(d, donation) {
  let row = document.createElement('tr');

  let contact = document.createElement('td');
  let contactLink = document.createElement('a');
  contact.appendChild(contactLink);

  if (donation == true) {
    contactLink.href = "javascript:window.c(" + d.donator + ");"
    contactLink.appendChild(smallCharacterImage(d.donator));
  } else {
    contactLink.href = "javascript:window.c(" + d.receiver + ");"
    contactLink.appendChild(smallCharacterImage(d.receiver));
  }

  let amount = document.createElement('td');
  amount.innerHTML = formatISK(d.amount);

  let note = document.createElement('td');
  note.innerHTML = d.note || '';

  let timestamp = document.createElement('td');
  timestamp.innerHTML = d.timestamp;

  row.appendChild(contact);
  row.appendChild(amount);
  row.appendChild(note);
  row.appendChild(timestamp);

  return row;
}

function contractRow(d, donation) {
  let row = document.createElement('tr');

  let contact = document.createElement('td');
  let contactLink = document.createElement('a');
  contact.appendChild(contactLink);

  if (donation == true) {
    contactLink.href = "javascript:window.c(" + d.donator + ");"
    contactLink.appendChild(smallCharacterImage(d.donator));
  } else {
    contactLink.href = "javascript:window.c(" + d.receiver + ");"
    contactLink.appendChild(smallCharacterImage(d.receiver));
  }

  let location = document.createElement('td');
  location.innerHTML = d.location;

  let system = document.createElement('td');
  system.innerHTML = d.system;

  let accepted = document.createElement('td');
  accepted.innerHTML = d.accepted;

  let issued = document.createElement('td');
  issued.innerHTML = d.issued;

  let expires = document.createElement('td');
  expires.innerHTML = d.expires;

  row.appendChild(contact);
  row.appendChild(location);
  row.appendChild(system);
  row.appendChild(accepted);
  row.appendChild(issued);
  row.appendChild(expires);

  return row;
}

function formatISK(n) {
  return n.toLocaleString(undefined, {
    maximumFractionDigits: 2,
    minimumFractionDigits: 2
  }) + ' ISK';
}

function largeCharacterImage(charID, charName, imgType, imgExtn) {
  let charImg = document.createElement('img');
  charImg.height = 150;
  charImg.width = 150;
  charImg.alt = charID.toString();
  charImg.classList.add('rounded');
  charImg.src = 'https://imageserver.eveonline.com/' + (imgType || 'Character') + '/' + charID + '_256.' + (imgExtn || 'jpg');

  let charDiv = document.createElement('div');
  let charNameP = document.createElement('p');
  charNameP.classList.add('text-muted');
  charNameP.classList.add('my-0');

  charNameP.innerHTML = charName;
  charDiv.appendChild(charImg);
  charDiv.appendChild(charNameP);

  return charDiv;
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

  jQuery.ajax({
    url: "/api/char?c=" + charID,
    success: function(t) {

      charImg.getElementsByTagName('p')[0].innerHTML = t.character.name;

      let corpImg = largeCharacterImage(
        t.character.corporation,
        t.character.corporation_name,
        'Corporation',
        'png',
      );
      corpImg.classList.add('col-2');
      character.appendChild(corpImg);

      if (t.donations != undefined) {
        details.appendChild(h3title("donations"));
        details.appendChild(donations);
        for (i=0; i < t.donations.length; i++) {
          donations.tBodies[0].appendChild(donationRow(t.donations[i], true));
        }
      }

      if (t.contracts != undefined) {
        details.appendChild(h3title("contracts"));
        details.appendChild(contracts);
        for (i=0; i < t.contracts.length; i++) {
          contracts.tBodies[0].appendChild(contractRow(t.contracts[i], true));
        }
      }

      if (t.donated != undefined) {

        details.appendChild(h3title("donated"));
        details.appendChild(donated);
        for (i=0; i < t.donated.length; i++) {
          donated.tBodies[0].appendChild(donationRow(t.donated[i], false));
        }
      }

      if (t.contracted != undefined) {
        details.appendChild(h3title("contracted"));
        details.appendChild(contracted);
        for (i=0; i < t.contracted.length; i++) {
          contracted.tBodies[0].appendChild(contractRow(t.contracted[i], false));
        }
      }

    },
    // XXX handle failure here
  });

  return details;
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

function content() {
  let container = document.createElement('div');
  container.id = "container";
  container.className = 'container-fluid';

  container.appendChild(signup());
  container.appendChild(header());

  var urlParams = new URLSearchParams(window.location.search);
  var charID = urlParams.get('c');

  if (charID != undefined) {
    container.appendChild(characterView(charID));
  } else {
    container.appendChild(frontPage());
  }

  container.appendChild(footer());
  return container;
}

document.body.removeChild(document.getElementById("nojs"));
document.body.appendChild(content());
window.c = switchCharacterView;
window.m = switchToMainPage;
