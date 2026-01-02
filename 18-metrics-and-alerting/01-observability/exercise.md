# Observability

In a couple of places in the previous modules, we mentioned that we are ignoring any potential incident handling and observability.
However, we have now come to the point where we need to learn more about methods that will help us understand what is going on in our system.

**This is a must-have when building a production-grade system.
We wanted to create a training that will give you enough knowledge to allow you to comfortably build event-driven systems.**
And that means we can't skip this topic.

If you are already knowledgeable about observability, some of what follows may be familiar to you.
We have tried to make it accessible for both people who don't have experience with this topic and those who already know something about it.

This training is not about observability, so we will not go extremely deep into that topic.
However, you should be comfortable with the topic after finishing this module.

## Why it's so important

**Basic monitoring of synchronous systems is quite easy: You can check whether the API is up, how many requests are being processed, and how many of those are failing.**
In some cases, you even don't need to modify your service to do that — you can have a proxy that will gather those metrics for you.

In event-driven systems, however, it's not that easy.
Just observing requests is not enough: Emitting the event may be successful, but that doesn't mean it will be processed.
**This is a trade-off related to improved fault tolerance and decoupling: Your system will be more resilient, but it will be harder to see when it is down.**
To overcome this, we will implement some basic metrics and suggest some basic alerts that you can add.
