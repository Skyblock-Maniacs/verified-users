import { createStyles } from "@mantine/styles"
import { getHotkeyHandler } from "@mantine/hooks"
import { ActionIcon, Center, Modal, SegmentedControl, SimpleGrid, TextInput } from "@mantine/core"
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
  const {classes, cx} = useStyles()

  const [modalOpen, setModalOpen] = useState(false)
  const [type, setType] = useState<"uuid" | "ign" | "discord">("uuid")
  const [value, setValue] = useState<string>("")
  const [user, setUserData] = useState<Response | null>(null)
  const [cloudflareToken, setCloudflareToken] = useState<string>("")

  const handleSubmit = async (token: string) => {
    await axios.post(`https://users.sbm.gg/api/v1/lookup/${type}/${value}`, {"cf-turnstile-response": token})
      .then(async (res) => {
        setUserData(res.data.data); 
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
        <TextInput
          placeholder="Search"
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
        />
        <SegmentedControl
          value={type}
          onChange={(value) => setType(value as "uuid" | "ign" | "discord")}
          data={[
            { label: "UUID", value: "uuid" },
            { label: "IGN", value: "ign" },
            { label: "Discord ID", value: "discord"}
          ]}
        />
        {
          user && (
            <SimpleGrid cols={2} spacing="md">
              <Center>
                <h1>UUID</h1>
                <p>{user.uuid}</p>
              </Center>
              <Center>
                <h1>Discord ID</h1>
                <p>{user.discordId}</p>
              </Center>
              <ReactSkinview3d
                height={500}
                width={500}
                skinUrl={`https://crafatar.com/skins/${user.uuid}`}
                capeUrl={`${user.cape}?notfromcache`}
              />
            </SimpleGrid>
          )
        }
      </div>
    </>
  )
}
