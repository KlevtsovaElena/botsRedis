let botsContainer = document.getElementById('bots_container')
let tmplBotsContainer = document.getElementById('tmpl-botsContainer').innerHTML
let tmplMessage = document.getElementById('tmpl-message').innerHTML
let botsHeader = document.querySelector('.bots_header_items')
let countContainers = botsHeader.querySelectorAll('.item')

    //функция для отправки запросов GET
function sendRequestGET(url){
    let requestObj = new XMLHttpRequest();
    requestObj.open('GET', url, false);
    requestObj.send();
    return requestObj.responseText;
}
renderBotsMessengers();
setInterval(function(){
    renderBotsMessengers()
}, 3000)


    //отрисуем полученные данные с помощью шаблона 
    function renderBotsMessengers(){
        botsContainer.innerHTML="";

        //запрос на получение api
        let json = sendRequestGET("http://localhost:8081/");

        //раскодируем данные
        let data = JSON.parse(json);

        console.log(data);
        let botsCount = data['BotsContents'].length

        let botsMessages = 0
        //заполним шаблоны
        for (let i = 0; i< botsCount; i++) {

            botsMessages += data['BotsContents'][i].Messages.length
           
            let messagesBot = "";
            for (let j = data['BotsContents'][i].Messages.length - 1; j >= 0; j-- ){
                
                //работаем с датой

                let dataTimeUser = new Date(data['BotsContents'][i].Messages[j].DateTime * 1000);

                let timeUser = dataTimeUser.toLocaleTimeString().slice(0, -3);

                //преобразуем дату с сервера в дату, которая у пользователя
                let dateUser = dataTimeUser.toLocaleDateString();


                //поработаем с юзернеймом
                //если нет username, то используем firtName, если нет и его, то lastName, если нет и его, то userId
                // if (data[i].Messages[j].Username = ""){
                //     if (data[i].Messages[j].FirstName != ""){

                //         if (data[i].Messages[j].LastName = "")
                //     }
                // }
                if (data['BotsContents'][i].Messages[j].IsImportant == 1){
                    messagesBot += tmplMessage.replace('class="message"', 'class="importantMessage"')
                                                .replace('${username}',data['BotsContents'][i].Messages[j].Username)
                                            .replace('${content}', data['BotsContents'][i].Messages[j].Content)
                                            .replace('${dateTime}', dateUser + " " + timeUser);
                } else {
                messagesBot += tmplMessage.replace('${username}',data['BotsContents'][i].Messages[j].Username)
                                            .replace('${content}', data['BotsContents'][i].Messages[j].Content)
                                            .replace('${dateTime}', dateUser + " " + timeUser);
                }
            }
            botsContainer.innerHTML += tmplBotsContainer.replace('${botName}', data['BotsContents'][i].Name)
                                                            .replace('${messages}', messagesBot);
        }

        countContainers[1].innerHTML = "Пользователей: " + data['UsersCount'];
        countContainers[0].innerHTML = "Ботов: " + botsCount;
        countContainers[2].innerHTML = "Сообщений: " + botsMessages;
                                        
    }
