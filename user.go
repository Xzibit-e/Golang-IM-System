package main

import (
	"net"
	"strings"
)

//2.创建用户类

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

//创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
		server: server,
	}

	//启动监听当前User channel消息的goroutine
	go user.ListenMessage()

	return user
}

//监听当前User channel的方法，一旦有消息，就直接发送给对端客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C

		u.conn.Write([]byte(msg + "\n"))
	}
}

//4. 对User类进行逻辑梳理和封装
func (u *User) Online() {
	//用户上线，将用户加入到onlineMap中
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	//广播当前用户上线消息
	u.server.BroadCast(u, "上线了！")
}

//4. 
func (u *User)Offline()  {
	//用户下线，将用户从onlineMap中删除
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	//广播当前用户下线消息
	u.server.BroadCast(u, "下线了！")
}

//4. 
func (u *User)DoMessage(msg string)  {
	//5. 用户处理消息的业务
	if msg == "who" {
		//6. 查询当前在线用户都有哪些

		u.server.mapLock.Lock() //!!!!!!!有上锁一定要解锁
		for _, user := range u.server.OnlineMap {
			onlionMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			u.SendMessage(onlionMsg)
		}
		u.server.mapLock.Unlock() //!!!!!!!!有上锁一定要解锁
	//7. 修改用户名
	} else if len(msg) >= 7 && msg[:7] == "rename|" {
		//7.1 消息格式：rename|张三
		newName := strings.Split(msg, "|")[1]

		//7.2 判断名称是否存在
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.SendMessage("当前名称已被占用！\n")
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()

			u.Name = newName
			u.SendMessage("已成功更新用户名：" + newName + "\n")
		}
	} else {
		u.server.BroadCast(u, msg)
	}
}

func (u *User)SendMessage(msg string)  {
	u.conn.Write([]byte(msg))
}