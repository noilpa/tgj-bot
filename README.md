## DESCRIPTION
Телеграм бот для атоматизации процесса ревью merge-requests

## FEATURES
- интеграция с Gitlab
- добавление участников через чат
- добавление merge-requests через чат
- равномерное распределение ревью между участниками
- рассылка напоминаний про ревью участникам (в общий чат)
- поддержка ролевой модели участников (developer и lead)
- поддержка состояний участника (active и inactive)
- механизм перераспределения ревью участника при смене статуса active --> inactive

## WORKFLOW
1. Зарегестрировать бота в телеграм у BotFather и заполнить конфиг-файл
2. Запустить бота
3. Добавить бота в чат
4. Зарегестрировать участников в боте: /register GitlabID role
5. Добавлять merge-requests: /mr url
6. При покидании проекта пользователь пишет: /inactive
7. При возвращении на проект пользователь пишет: /active

## DEPLOY
Скачать проект и собрать контейнер
```bash
git clone tgj-bot-repo-url
cd tgj-bot
docker build .
```
Подготовить конфиг
```bash
nano config/config.json
```
Запустить образ (image_id получается после выполнения docker build .)
```bash
docker run -v $(pwd)/db:/db -v $(pwd)/conf:/conf -d image_id```

------------


В общем виде команда выглядит
```bash
docker run -v /host/dir/db:/db /host/dir/conf:/conf -d imaje_id
```