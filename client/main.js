(() => {
    fetch("/api/v1/hitokoto").then(res => {
        return res.json();
    }).then(json => {
        const hitokoto = document.getElementById("hitokoto");
        hitokoto.innerText = `『${json.Data.Content}』 —— ${json.Data.User.Name}`;
    });

    const ap = new APlayer({
        container: document.getElementById("music"),
        audio: [{
            name: '暗號',
            artist: 'Jay Chou',
            url: 'music.mp3',
            cover: 'cover.webp'
        }]
    });
})();
