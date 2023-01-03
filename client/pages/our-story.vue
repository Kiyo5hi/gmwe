<template>
  <div class="flex flex-col items-center justify-center">
    <ul class="steps steps-vertical">
      <li v-for="event in events" :key="event.event" data-content="●" class="step step-primary">
        {{ event.date }}: {{ event.event }}
      </li>
    </ul>
  </div>
</template>

<script setup>
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
  date: '2022-08-07',
  event: 'First Ring'
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
}].sort((a, b) => a.date.localeCompare(b.date))
</script>
