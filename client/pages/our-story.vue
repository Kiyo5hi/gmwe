<template>
  <section class="story-page">
    <header><h1>Story</h1><p>Acey &amp; Kiyoshi · {{ events.length }} 个瞬间</p></header>
    <ol class="story-timeline">
      <li v-for="event in events" :key="event.date + event.event">
        <time :datetime="event.date">{{ event.date.replaceAll('-', '.') }}</time>
        <p>{{ event.event }}</p>
      </li>
    </ol>
  </section>
</template>

<script lang="ts" setup>
const events = [{
  date: '2022-06-18',
  event: '初识'
}, {
  date: '2022-06-27',
  event: '初见'
}, {
  date: '2022-07-06',
  event: 'Kiyoshi 纹身'
}, {
  date: '2022-07-11',
  event: '狗美甲'
}, {
  date: '2022-07-19',
  event: '开始'
}, {
  date: '2022-07-26',
  event: '吵架'
}, {
  date: '2022-07-27',
  event: '纹身'
}, {
  date: '2022-08-04',
  event: '第一次七夕节'
}, {
  date: '2022-08-07',
  event: 'First Ring'
}, {
  date: '2022-09-11',
  event: '奇妙的一天 and 橡木郡'
}, {
  date: '2022-09-12',
  event: '电子化'
}, {
  date: '2022-10-06',
  event: '鸭给狗买的秋装和鸭信 Batch 1 签收'
}, {
  date: '2022-10-08',
  event: '狗信 Batch 1 签收'
}, {
  date: '2022-10-22',
  event: '鸭信 Batch 2 和牵引绳签收'
}, {
  date: '2022-10-25',
  event: '狗信 Batch 2 签收'
}, {
  date: '2022-10-26',
  event: '100 Days'
}, {
  date: '2022-11-18',
  event: '很重要的电子谈心'
}, {
  date: '2022-12-15',
  event: '鸭信 Batch 3 签收'
}, {
  date: '2022-12-15',
  event: '狗信 Batch 3 签收'
}, {
  date: '2022-12-27',
  event: '鸭给狗的圣诞礼物和生日信签收'
}, {
  date: '2023-01-01',
  event: '新篇章'
}, {
  date: '2023-02-03',
  event: '200 Days'
}, {
  date: '2023-02-14',
  event: 'First Valentine\'s Day'
}, {
  date: '2023-02-19',
  event: '鸭狗围巾'
}, {
  date: '2023-06-19',
  event: '狗归狗回'
}, {
  date: '2023-06-23',
  event: '那不勒斯意面，灌，男友暂存'
}, {
  date: '2023-07-08',
  event: 'Second Ring'
}, {
  date: '2023-07-15',
  event: '西拍答谢宴'
}, {
  date: '2023-08-01',
  event: '鸭袜 Replication'
}, {
  date: '2023-08-05',
  event: '分手体验卡'
}, {
  date: '2023-08-09',
  event: '分手体验卡到期'
}, {
  date: '2023-08-16',
  event: '小动物们的表情包'
}, {
  date: '2023-08-18',
  event: '巨大鸭袜到达'
}, {
  date: '2023-08-19',
  event: 'IKEA'
}, {
  date: '2023-08-20',
  event: '给 Aron 打猎'
}, {
  date: '2023-08-22',
  event: '重游橡木郡'
}, {
  date: '2023-08-27',
  event: '小动物 Tony'
}, {
  date: '2023-09-06',
  event: '🐶得到第一份工作'
}, {
  date: '2023-09-09',
  event: 'Observatory 🌃'
}, {
  date: '2023-09-09',
  event: '🐱 Aron 6.8'
}].sort((a, b) => a.date.localeCompare(b.date))
</script>
<style scoped>
.story-page { max-width: 720px; margin: 0 auto; }
header > p { margin-top: 10px; color: var(--muted); font-size: 13px; }
.story-timeline { list-style: none; margin: 40px 0 0; padding: 0; }
li { display: grid; grid-template-columns: 108px minmax(0, 1fr); gap: 32px; position: relative; padding-bottom: 30px; }
time { color: var(--muted); font-size: 12px; padding-top: 4px; font-variant-numeric: tabular-nums; }
li p { font-size: 16px; line-height: 1.7; padding-left: 22px; border-left: 1px solid var(--line); position: relative; }
li p::before { content: ''; width: 7px; height: 7px; background: var(--accent); border-radius: 50%; position: absolute; left: -4px; top: 9px; }
@media (max-width: 480px) { li { grid-template-columns: 86px minmax(0, 1fr); gap: 10px; padding-bottom: 24px; } li p { font-size: 15px; padding-left: 16px; } }
</style>
