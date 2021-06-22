# Запуск проекта в Kubernetes за 60 минут

## Создание инфраструктуры в облаке Mail.ru Cloud Solutions

1. Регистрируемся в [Mail.ru Cloud Solutions](https://mcs.mail.ru)
1. Заходим в личный кабинет MCS и создаем кластер Kubernetes.
В предустановленных сервисах оставляем Nginx Ingress Controller
1. После создания скачиваем kubeconfig файл и экспортируем его

```bash
export KUBECONFIG=<path to downloaded kubeconfig>
```

1. Пока создается кластер, создадим базу данных
  1. Выбираем PostgreSQL
  1. На этапе создания пользователя, жмем автосгенерировать пароль и сохраняем
  его, он нам понадобится в дальнейшем

## Создание проекта в Gitlab.com

1. Регистрируемся на [Gitlab.com](https://gitlab.com)
1. После регистрации создаем новый пустой проект с именем mcs-kubernetes-project
1. Клонируем его себе
1. Копируем содержимое директории app из текущего репозитория в репозиторий гитлаба
1. Пушим изменения

## Добавление раннера

1. Добавляем хелм репозиторий гитлаба

```bash
helm repo add gitlab https://charts.gitlab.io
```

1. Ставим раннер

```bash
kubectl create namespace gitlab
helm install --namespace gitlab gitlab-runner gitlab/gitlab-runner \
  --set rbac.create=true \
  --set runners.privileged=true \
  --set gitlabUrl=https://gitlab.com/ \
  --set runnerRegistrationToken=<token from project settings/CI/CD/Runners>
```

1. Выключаем шаренные раннеры в Settings/CI/CD/Runners

## Создание CI

1. Копируем файл .gitlab-ci.yml из текущего репозитория
в репозиторий на гитлабе и пушим изменения
1. Смотрим за пайплайном в интерфейсе гитлаба. Шаг деплоя на данном этапе
должен падать, так и задумано

## Подготовка Kubernetes к деплою

1. Создаем неймспейс для приложения

```bash
kubectl create ns mcs-kubernetes-project
```

1. Создаем RBAC объекты для доступа раннера к API Kubernetes

```bash
kubectl create sa deploy -n mcs-kubernetes-project
kubectl create rolebinding deploy \
  -n mcs-kubernetes-project \
  --clusterrole edit \
  --serviceaccount mcs-kubernetes-project:deploy
```

1. Получам токен от созданного сервисаккаунта

```bash
kubectl get secret -n mcs-kubernetes-project \
  $(kubectl get sa -n mcs-kubernetes-project deploy \
    -o jsonpath='{.secrets[].name}') \
  -o jsonpath='{.data.token}'
```

1. Переходим в интерфейс гитлаба. В левом меню находим Settings, далее CI/CD
и далее Variables и нажимаем Expand
В левое поле вводим имя переменной
K8S_CI_TOKEN
В правое поле вводим скопированный токен из вывода команды setup.sh
Protected выключаем!
Masked включаем!
1. Далее в том же левом меню в Settings > Repository находим Deploy tokens и нажимаем Expand.
В поле Name вводим
k8s-pull-token
И ставим галочку рядом с read_registry.
Все остальные поля оставляем пустыми.
Нажимаем Create deploy token.
НЕ ЗАКРЫВАЕМ ОКНО БРАУЗЕРА!
1. Возвращаемся в консоль.
Создаем image pull secret для того чтобы кубернетис мог пулить имаджи из гитлаба

```bash
kubectl create secret docker-registry mcs-kubernetes-project-image-pull \
  --docker-server registry.gitlab.com \
  --docker-email 'admin@mycompany.com' \
  --docker-username '<первая строчка из окна создания токена в gitlab>' \
  --docker-password '<вторая строчка из окна создания токена в gitlab>' \
  --namespace mcs-kubernetes-project
```

Соответсвенно подставляя на место <> нужные параметры.

1. Создаем секрет с паролем к БД

```bash
kubectl create secret generic mcs-kubernetes-project \
  -n mcs-kubernetes-project \
  --from-literal db-password=<сохраненный ранее пароль от БД>
```

## Подготовка манифестов Kubernetes для проекта

1. Копируем директорию .kube из текущего репозитория в репозиторий проекта на гитлабе
1. Проставляем в файле deployment.yaml параметры подключения к БД
1. Пушим изменения, ждем окончания деплоя
1. По IP адресу балансировщика созданного облаком проверяем работу приложения.
Его можно получить с помощью команды

```bash
kubectl get svc -n ingress-nginx nginx-ingress-controller -o jsonpath='{.status.loadBalancer.ingress[].ip}'
```

1. Выполняем пару запросов для проверки

```bash
curl -k https://<INGRESS IP>/users -X POST -H "Content-Type: application/json" --data '{"name":"test","location":"asdf","age":12}'
curl -k https://<INGRESS IP>/users
```
