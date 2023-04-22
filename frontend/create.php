<?php

var_dump($_POST);

//получаем данные из формы
$title = $_POST["title"];
$time = $_POST["time"];
$user_id = intval($_POST["user_id"]);
$description = $_POST["description"];

//преобразуем время в секунды
$timeSeconds = strtotime($time);


//создаем соединение с БД
$host = 'mysql';
$db   = 'inordic';
$user = 'root';
$pass = 'nordic123';
$charset = 'utf8';

$dsn = "mysql:host=$host;dbname=$db;charset=$charset";
$opt = [
        PDO::ATTR_ERRMODE            => PDO::ERRMODE_EXCEPTION,
        PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
        PDO::ATTR_EMULATE_PREPARES   => false,
];
//создание объекта подключения к БД
$pdo = new PDO($dsn, $user, $pass, $opt);

//формируем запрос
$sqlText = "INSERT INTO `tasks`(`title`,`time`,`user_id`,`description`) VALUES(:title, :time, :user_id, :description) ";

//подготавливаем запрос
$stmt = $pdo->prepare($sqlText);
//проставляем данные в пропуски в запросе
$stmt->bindParam(":title",$title);
$stmt->bindParam(":time",$timeSeconds);
$stmt->bindParam(":user_id",$user_id);
$stmt->bindParam(":description",$description);
//отправляем запрос в базу данных
$stmt->execute();
