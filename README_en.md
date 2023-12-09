# arena

Machine translation into English

#### Introduction

Go1.20 introduces arena. Arena, but it cannot be used by multiple goroutines at the same time, and go1.n (n<20) cannot be used. This package provides Arena to remove these restrictions

#### Features

- It can be used in multiple goroutines at the same time

- Do not use generics. You can use generics in versions earlier than go1.20
