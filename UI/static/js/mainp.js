const imageWrapper = document.querySelector('.highlight-image-wrapper');
const image = document.querySelector('.highlight-image');

imageWrapper.addEventListener('mousemove', (e) => {
    const xAxis = (window.innerWidth / 2 - e.pageX) / 25;
    const yAxis = (window.innerHeight / 2 - e.pageY) / 25;
    image.style.transform = `rotateY(${xAxis}deg) rotateX(${yAxis}deg)`;
});

imageWrapper.addEventListener('mouseenter', () => {
    image.style.transition = 'none'; /* Убираем анимацию при наведении */
});

imageWrapper.addEventListener('mouseleave', () => {
    image.style.transition = 'transform 0.5s ease';
    image.style.transform = `rotateY(0deg) rotateX(0deg)`; /* Возвращаем в исходное положение */
});

// Mapbox initialization
mapboxgl.accessToken = 'pk.eyJ1Ijoicm9uaW4zNTQ2NTciLCJhIjoiY20wdXAxZG9nMTdtajJpczVkYnlhN2c0YiJ9.vAUFyu6IRiMEZSlhfDnj8A'; // Убедитесь, что вы используете действительный токен
const map = new mapboxgl.Map({
    container: 'map', // контейнер для карты
    style: 'mapbox://styles/mapbox/streets-v11', // стиль карты
    center: [37.618423, 55.751244], // начальные координаты (центр Москвы)
    zoom: 2 // начальный зум
});

// Добавление точек серверов (пример)
const servers = [
    { coordinates: [37.618423, 55.751244], name: 'Москва' },
    { coordinates: [-0.1276, 51.5074], name: 'Лондон' },
    { coordinates: [-74.006, 40.7128], name: 'Нью-Йорк' },
    { coordinates: [13.4105, 52.5244], name: 'Берлин' },
    { coordinates: [4.88969, 52.374], name: 'Амстердам' },
];

servers.forEach(server => {
    new mapboxgl.Marker()
        .setLngLat(server.coordinates)
        .setPopup(new mapboxgl.Popup().setHTML(`<h3>${server.name}</h3>`))
        .addTo(map);
});