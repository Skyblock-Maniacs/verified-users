import { createStyles } from "@mantine/styles"
import { getHotkeyHandler } from "@mantine/hooks"
import { ActionIcon, Center, Modal, SimpleGrid, TextInput, Text, Title, Paper, Avatar, Image, ColorSwatch, Flex } from "@mantine/core"
import { notifications } from '@mantine/notifications';
import { useState } from "react"
import { SquareArrowRight } from "tabler-icons-react"
import axios from "axios"
import Turnstile from "react-turnstile"
import dynamic from "next/dynamic";

const ReactSkinview3d = dynamic(() => import("react-skinview3d"), { ssr: false });

const useStyles = createStyles((theme) => ({
  container: {
    display: "flex",
    justifyContent: "center",
    padding: "1rem",
    flexDirection: "column",
    alignItems: "center"
  },
  checkbox: {
    minHeight: "70px"
  },
  input: {
    width: "40%"
  },
  title: {
    "-webkit-user-select": "none",
    "-ms-user-select": "none",
    "user-select": "none"
  },
  discordContainer: {
    display: "flex",
    justifyContent: "start-left",
    padding: "1rem",
    flexDirection: "row",
    alignItems: "center",
    gap: "3rem"
  },
  colorSwatch: {
    width: "70px",
    height: "25px"
  }
}))

interface Response {
  uuid: string;
  discordId: string;
  ign: string;
  skin: string;
  cape: string;
  discordUser: {
    accent_color: number;
    avatar: string;
    banner: string;
    banner_color: string;
    bot: boolean
    created_at: number;
    discriminator: string;
    id: string;
    public_flags: string;
    system: boolean;
    username: string
  }
}

export default function Home() {
  const { classes } = useStyles()

  const [modalOpen, setModalOpen] = useState(false)
  const [value, setValue] = useState<string>("")
  const [user, setUserData] = useState<Response | null>(null)

  const handleSubmit = async (token: string) => {
    let type: "uuid" | "ign" | "discord"
    if (3 <= value.length && value.length <= 16) {
      type = "ign"
    } else if (17 <= value.length && value.length <= 19) {
      type = "discord"
    } else if (32 <= value.length && value.length <= 36) {
      type = "uuid"
      setValue(value.replace(/-/g, ""))
    } else {
      notifications.show({
        title: "Error",
        message: "Invalid input length",
        autoClose: 10000,
        color: "red"
      })
      setModalOpen(false)
      return
    }
    await axios.post(`https://users.sbm.gg/api/v1/lookup/${type}/${value}`, {"cf-turnstile-response": token})
      .then((res) => {
        console.log(res.data.data)
        setUserData(res.data.data); 
        console.log(user)
        setModalOpen(false)}
      )
      .catch((err) => {
        console.log(err)
        setModalOpen(false)
        notifications.show({
          title: "Error",
          message: err.response.data.message,
          autoClose: 10000,
          color: "red"
        })
      })
  }

  return (
    <>
      <Modal
        opened={modalOpen}
        onClose={() => setModalOpen(false)}
        title="Captcha Required"
        size="auto"
        centered={true}
        closeOnEscape={false}
        closeOnClickOutside={false}
      >
        <Turnstile
          sitekey={process.env.NEXT_PUBLIC_CLOUDFLARE_TURNSTILE_SITE_KEY as string}
          onVerify={handleSubmit}
        />
      </Modal>
      <div className={classes.container}>
        <Title
          color="white"
          sx={{ fontFamily: 'Greycliff CF, sans-serif' }}
          className={classes.title}
          size={54}
          mb={32}
        >
          Hypixel Skyblock Verified Users
        </Title>
        <TextInput
          placeholder="Enter a username, UUID, or Discord ID"
          value={value}
          onChange={(e) => setValue(e.currentTarget.value)}
          rightSection={
            <ActionIcon onClick={() => setModalOpen(true)}><SquareArrowRight/></ActionIcon>
          }
          onKeyDown={
            getHotkeyHandler([
              ['Enter', () => setModalOpen(true)]
            ])
          }
          className={classes.input}
          label={<Text size={"md"} color="white">IGN/UUID/Discord ID:</Text>}
          radius="lg"
        />
        {
          user && (
            <SimpleGrid cols={2} spacing="md" mt={50}>
              <Paper shadow="md" p="md" bg="#424549">
                <Title align="center" color="white" mb="sm" underline>Minecraft Information</Title>
                <SimpleGrid cols={2} spacing="md">
                  <Paper p="md" bg="#36393e">
                    <ReactSkinview3d
                      height={500}
                      width={320}
                      skinUrl={`https://crafatar.com/skins/${user.uuid}`}
                      capeUrl={user.cape != "" ? user.cape : undefined}
                    />
                  </Paper>
                  <div>
                    <Text>UUID: {user.uuid}</Text>
                    <Text>IGN: {user.ign}</Text>
                  </div>
                </SimpleGrid>
              </Paper>
              <Paper shadow="md" p="md" bg="#424549">
                <Title align="center" color="white" mb="sm" underline>Discord Information</Title>
                <Paper p="md" bg="#36393e">
                  <div className={classes.discordContainer}>
                    <Avatar 
                      src={`https://cdn.discordapp.com/avatars/${user.discordId}/${user.discordUser.avatar}?size=1024`} 
                      size={100} 
                      radius={100}
                    />
                    <Image 
                      src={`https://cdn.discordapp.com/banners/${user.discordId}/${user.discordUser.banner}?size=1024`} 
                      maw={375}
                      radius={5}
                    />
                  </div>
                  <Text>ID: {user.discordId}</Text>
                  <Text># Username: {user.discordUser.username}#{user.discordUser.discriminator}</Text>
                  <Text>Created: {new Date(user.discordUser.created_at * 1000).toUTCString()}</Text>
                  <Flex gap="sm">
                    <Text>Banner Color:</Text>
                    <ColorSwatch 
                      color={user.discordUser.banner_color}
                      radius={30}
                      className={classes.colorSwatch}
                      withShadow={false}
                    />
                  </Flex>
                </Paper>
              </Paper>
            </SimpleGrid>
          )
        }
      </div>
    </>
  )
}
