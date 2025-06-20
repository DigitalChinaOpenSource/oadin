!(function (e, t) {
  e.WxLogin = function (n) {
    var r = 'default'
    !0 === n.self_redirect ? (r = 'true') : !1 === n.self_redirect && (r = 'false')
    var o = t.createElement('iframe'),
      s =
        'https://open.weixin.qq.com/connect/qrconnect?appid=' +
        n.appid +
        '&scope=' +
        n.scope +
        '&redirect_uri=' +
        n.redirect_uri +
        '&state=' +
        n.state +
        '&login_type=jssdk&self_redirect=' +
        r +
        '&styletype=' +
        (n.styletype || '') +
        '&sizetype=' +
        (n.sizetype || '') +
        '&bgcolor=' +
        (n.bgcolor || '') +
        '&rst=' +
        (n.rst || '')
    ;(s += n.style ? '&style=' + n.style : ''),
      (s += n.href ? '&href=' + n.href : ''),
      (s += 'en' === n.lang ? '&lang=en' : ''),
      (s += 1 === n.stylelite ? '&stylelite=1' : ''),
      (s += 0 === n.fast_login ? '&fast_login=0' : ''),
      (o.src = s),
      (o.frameBorder = '0'),
      (o.allowTransparency = 'true'),
      (o.scrolling = 'no'),
      (o.width = '160px'),
      (o.height = '160px')
    var i = t.getElementById(n.id)
    if (((i.innerHTML = ''), i.appendChild(o), e.addEventListener && e.JSON && n.onReady && 'function' == typeof n.onReady)) {
      var a = function (t) {
        if ('https://open.weixin.qq.com' === t.origin)
          try {
            var r = JSON.parse(t.data)
            if (r && 'status' === r.type) {
              var o = 'wxReady' === r.status
              o && n.onReady(o)
            }
          } catch (t) {
            e.console && 'function' == typeof e.console.log && e.console.log('wxLogin postMessage error', t)
          }
      }
      e.addEventListener('message', a, !1)
      var l = !1
      n.onCleanup = function () {
        !l && e.removeEventListener && (e.removeEventListener('message', a, !1), (l = !0))
      }
    }
  }
})(window, document)
