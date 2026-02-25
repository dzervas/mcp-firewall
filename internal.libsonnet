// strict normalization to internal representation
local allowedKeys = { allow: true, ask: true, deny: true };

local assertRule(name, r) =
  assert std.type(name) == 'string' && name != '' : 'rule name must be non-empty string';
  assert std.type(r) == 'object' : 'rule ' + name + ' must be object';
  assert std.length([k for k in std.objectFields(r) if !std.objectHas(allowedKeys, k)]) == 0 :
         'rule ' + name + ' contains unknown keys';
  assert std.type(if std.objectHas(r, 'allow') then r.allow else []) == 'array' : 'rule ' + name + '.allow must be array';
  assert std.type(if std.objectHas(r, 'ask') then r.ask else []) == 'array' : 'rule ' + name + '.ask must be array';
  assert std.type(if std.objectHas(r, 'deny') then r.deny else []) == 'array' : 'rule ' + name + '.deny must be array';
  assert std.length([x for x in (if std.objectHas(r, 'allow') then r.allow else []) if std.type(x) != 'string' || std.trim(x) == '']) == 0 :
         'rule ' + name + '.allow must contain non-empty strings';
  assert std.length([x for x in (if std.objectHas(r, 'ask') then r.ask else []) if std.type(x) != 'string' || std.trim(x) == '']) == 0 :
         'rule ' + name + '.ask must contain non-empty strings';
  assert std.length([x for x in (if std.objectHas(r, 'deny') then r.deny else []) if std.type(x) != 'string' || std.trim(x) == '']) == 0 :
         'rule ' + name + '.deny must contain non-empty strings';
  {
    allow: if std.objectHas(r, 'allow') then r.allow else [],
    ask: if std.objectHas(r, 'ask') then r.ask else [],
    deny: if std.objectHas(r, 'deny') then r.deny else [],
  };

function(rules)
  assert std.type(rules) == 'object' : 'rules must be an object';
  { [name]: assertRule(name, rules[name]) for name in std.objectFields(rules) }
